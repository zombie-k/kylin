package pool

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

// Slice.
type Slice struct {
	// New is an application supplied function for creating and
	// configuring a item.
	New func(ctx context.Context) (io.Closer, error)
	// stop cancel the item opener.
	stop func()

	// mutex protects fields defined below.
	mutex        sync.Mutex
	freeItem     []*item
	itemRequests map[uint64]chan item
	nextRequest  uint64 // next key use in itemRequests
	active       int    // number of opened and pending open items
	// Used to signal the need for new items.
	// a goroutine running itemOpener() reads on this chan and
	// maybeOpenNewItems sends on the chan (one send per needed item)
	// It is closed during db.Close(). The close tells the itemOpener
	// goroutine to exit.
	openerCh  chan struct{}
	closed    bool
	cleanerCh chan struct{}

	// Config pool configuration
	conf *Config
}

func NewSlice(c *Config) *Slice {
	// check config
	if c == nil {
		panic("config nil")
	}
	if c.Active < c.Idle {
		panic("Idle must <= Active")
	}
	ctx, cancel := context.WithCancel(context.Background())
	// new pool
	p := &Slice{
		conf:         c,
		stop:         cancel,
		itemRequests: make(map[uint64]chan item),
		openerCh:     make(chan struct{}, 1000000),
	}
	p.startCleanerLocked(time.Duration(c.IdleTimeout))

	go p.itemOpener(ctx)
	return p
}

func (p *Slice) Get(ctx context.Context) (io.Closer, error) {
	p.mutex.Lock()
	if p.closed {
		p.mutex.Unlock()
		return nil, ErrPoolClosed
	}
	idleTimeout := time.Duration(p.conf.IdleTimeout)
	// Prefer a free item if exist.
	numFree := len(p.freeItem)
	for numFree > 0 {
		i := p.freeItem[0]
		copy(p.freeItem, p.freeItem[1:])
		p.freeItem = p.freeItem[:numFree-1]
		p.mutex.Unlock()
		if i.expired(idleTimeout) {
			i.close()
			p.mutex.Lock()
			p.release()
		} else {
			return i.c, nil
		}
		numFree = len(p.freeItem)
	}

	// Out of free items or we were asked not to use one. If we're not
	// allowed to open any more items, make a request and wait.
	if p.conf.Active > 0 && p.active >= p.conf.Active {
		// check waitTimeout and return directory
		if p.conf.WaitTimeout == 0 && !p.conf.Wait {
			p.mutex.Unlock()
			return nil, ErrPoolExhausted
		}

		// Make the item channel. It's buffered so that the itemOpener
		// doesn't block while waiting for the req to be read.
		req := make(chan item, 1)
		reqKey := p.nextRequestKeyLocked()
		p.itemRequests[reqKey] = req
		wt := p.conf.WaitTimeout
		p.mutex.Unlock()

		// reset context timeout
		if wt > 0 {
			var cancel func()
			_, ctx, cancel = wt.Shrink(ctx)
			defer cancel()
		}

		// Timeout the item request with the context
		select {
		case <-ctx.Done():
			// Remove the item request and ensure no value has been sent
			// on it after removing.
			p.mutex.Lock()
			delete(p.itemRequests, reqKey)
			p.mutex.Unlock()
			return nil, ctx.Err()
		case ret, ok := <-req:
			if !ok {
				return nil, ErrPoolClosed
			}
			if ret.expired(idleTimeout) {
				ret.close()
				p.mutex.Lock()
				p.release()
			} else {
				return ret.c, nil
			}
		}
	}

	p.active++
	p.mutex.Unlock()
	c, err := p.New(ctx)
	if err != nil {
		p.mutex.Lock()
		p.release()
		p.mutex.Unlock()
		return nil, err
	}
	return c, nil
}

func (p *Slice) Put(ctx context.Context, c io.Closer, forceClose bool) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if forceClose {
		p.release()
		return c.Close()
	}
	added := p.putItemLocked(c)
	if !added {
		p.active--
		return c.Close()
	}
	return nil
}

// itemOpener runs in a separate goroutine, opens new item when requested.
func (p *Slice) itemOpener(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.openerCh:
			p.openNewItem(ctx)
		}
	}
}

func (p *Slice) maybeOpenNewItems() {
	numRequests := len(p.itemRequests)
	if p.conf.Active > 0 {
		numCanOpen := p.conf.Active - p.active
		if numRequests > numCanOpen {
			numRequests = numCanOpen
		}
	}
	for numRequests > 0 {
		p.active++
		numRequests--
		if p.closed {
			return
		}
		p.openerCh <- struct{}{}
	}
}

// openNewItem one new item.
func (p *Slice) openNewItem(ctx context.Context) {
	c, err := p.New(ctx)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if err != nil {
		p.release()
	}
	if !p.putItemLocked(c) {
		p.active--
		c.Close()
	}
}

func (p *Slice) putItemLocked(c io.Closer) bool {
	if p.closed {
		return false
	}
	if p.conf.Active > 0 && p.active > p.conf.Active {
		return false
	}
	i := item{
		createAt: nowFunc(),
		c:        c,
	}
	if l := len(p.itemRequests); l > 0 {
		var req chan item
		var reqKey uint64
		for reqKey, req = range p.itemRequests {
			break
		}
		delete(p.itemRequests, reqKey) // remove from pending requests.
		req <- i
		return true
	} else if !p.closed && p.maxIdleItemsLocked() > len(p.freeItem) {
		p.freeItem = append(p.freeItem, &i)
		return true
	}
	return false
}

// startCleanerLocked starts itemCleaner if needed.
func (p *Slice) startCleanerLocked(d time.Duration) {
	if d <= 0 {
		return
	}
	if d < time.Duration(p.conf.IdleTimeout) && p.cleanerCh != nil {
		select {
		case p.cleanerCh <- struct{}{}:
		default:
		}
	}
	// run only one, clean stale items.
	if p.cleanerCh == nil {
		p.cleanerCh = make(chan struct{}, 1)
		go p.staleCleaner(time.Duration(p.conf.IdleTimeout))
	}
}

func (p *Slice) staleCleaner(d time.Duration) {
	const minInterval = 100 * time.Millisecond
	if d < minInterval {
		d = minInterval
	}
	t := time.NewTimer(d)
	for {
		select {
		case <-t.C:
		case <-p.cleanerCh: //maxLifetime was changed or db was closed.
		}
		p.mutex.Lock()
		d = time.Duration(p.conf.IdleTimeout)
		if p.closed || d <= 0 {
			p.mutex.Unlock()
			return
		}
		expiredSince := nowFunc().Add(-d)
		var closing []*item
		for i := 0; i < len(p.freeItem); i++ {
			c := p.freeItem[i]
			if c.createAt.Before(expiredSince) {
				closing = append(closing, c)
				p.active--
				last := len(p.freeItem) - 1
				p.freeItem[i] = p.freeItem[last]
				p.freeItem[last] = nil
				p.freeItem = p.freeItem[:last]
				i--
			}
		}
		p.mutex.Unlock()
		for _, c := range closing {
			c.close()
		}
		if d < minInterval {
			d = minInterval
		}
		t.Reset(d)
	}
}

func (p *Slice) nextRequestKeyLocked() uint64 {
	next := p.nextRequest
	p.nextRequest++
	return next
}

const defaultIdleItems = 2

func (p *Slice) maxIdleItemsLocked() int {
	n := p.conf.Idle
	switch {
	case n == 0:
		return defaultIdleItems
	case n < 0:
		return 0
	default:
		return n
	}
}

func (p *Slice) release() {
	p.active--
	p.maybeOpenNewItems()
}

func (p *Slice) Close() error {
	p.mutex.Lock()
	if p.closed {
		p.mutex.Unlock()
		return nil
	}
	if p.cleanerCh != nil {
		close(p.cleanerCh)
	}
	var err error
	for _, i := range p.freeItem {
		i.close()
	}
	p.freeItem = nil
	p.closed = true
	for _, req := range p.itemRequests {
		close(req)
	}
	p.mutex.Unlock()
	p.stop()
	return err
}

func (p *Slice) String() string {
	return fmt.Sprintf("freeItem:%+v itemRequests:%+v nextRequest:%d active:%d mutex:%+v", p.freeItem, p.itemRequests, p.nextRequest, p.active, p.mutex)
}

package fanout

import (
	"context"
	"errors"
	"github.com/zombie-k/kylin/library/log"
	"runtime"
	"sync"
)

var (
	ErrFull = errors.New("fanout: chan full")
)

type options struct {
	worker int
	buffer int
}

type Option func(*options)

func Worker(n int) Option {
	if n <= 0 {
		panic("fanout: worker should > 0")
	}
	return func(o *options) {
		o.worker = n
	}
}

func Buffer(n int) Option {
	if n <= 0 {
		panic("fanout: buffer should > 0")
	}
	return func(o *options) {
		o.buffer = n
	}
}

type item struct {
	f   func(c context.Context)
	ctx context.Context
}

type Fanout struct {
	name    string
	ch      chan item
	options *options
	wg      sync.WaitGroup

	ctx    context.Context
	cancel func()
}

func (f *Fanout) Channel() int {
	return len(f.ch)
}

func New(name string, opts ...Option) *Fanout {
	if name == "" {
		name = "anonymous"
	}
	o := &options{
		worker: 1,
		buffer: 1024,
	}
	for _, op := range opts {
		op(o)
	}
	f := &Fanout{
		name:    name,
		ch:      make(chan item, o.buffer),
		options: o,
		wg:      sync.WaitGroup{},
	}
	f.ctx, f.cancel = context.WithCancel(context.Background())
	f.wg.Add(o.worker)
	for i := 0; i < o.worker; i++ {
		go f.proc()
	}
	return f
}

func (f *Fanout) proc() {
	defer f.wg.Done()
	for {
		select {
		case t := <-f.ch:
			wrapFunc(t.f)(t.ctx)
		case <-f.ctx.Done():
			return
		}
	}
}

func wrapFunc(fn func(ctx context.Context)) (res func(ctx context.Context)) {
	res = func(ctx context.Context) {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64*1024)
				buf = buf[:runtime.Stack(buf, false)]
				log.Error("panic in fanout proc, err: %s, stack: %s", r, buf)
			}
		}()
		fn(ctx)
	}
	return
}

func (f *Fanout) Do(ctx context.Context, fn func(ctx context.Context)) (err error) {
	if fn == nil || f.ctx.Err() != nil {
		return f.ctx.Err()
	}
	select {
	case f.ch <- item{f: fn, ctx: ctx}:
	default:
		err = ErrFull
	}
	return
}

// DoWait save a callback func, blocking when channel is full.
func (f *Fanout) DoWait(ctx context.Context, fn func(ctx context.Context)) (err error) {
	if fn == nil || f.ctx.Err() != nil {
		return f.ctx.Err()
	}
	select {
	case f.ch <- item{f: fn, ctx: ctx}:
	}
	return
}

func (f *Fanout) Close() error {
	if err := f.ctx.Err(); err != nil {
		return err
	}
	f.cancel()
	f.wg.Wait()
	return nil
}

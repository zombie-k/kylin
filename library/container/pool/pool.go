package pool

import (
	"context"
	"errors"
	xtime "github.com/zombie-k/kylin/library/time"
	"io"
	"time"
)

var (
	// ErrPoolExhausted connections are exhausted.
	ErrPoolExhausted = errors.New("container/pool exhausted")
	// ErrPoolClosed connection pool is closed.
	ErrPoolClosed = errors.New("container/pool closed")

	nowFunc = time.Now
)

// Pool interface.
type Pool interface {
	Get(ctx context.Context) (io.Closer, error)
	Put(ctx context.Context, c io.Closer, forceClose bool) error
	Close() error
}

// Config is the pool configuration struct.
type Config struct {
	// Active number of items allocated by the pool at a given time.
	// When zero, there is no limit on the number of items in the pool.
	Active int
	// Idle number of idle items in the pool.
	Idle int
	// Close items after remaining item for this duration. If the value
	// is zero, then the items are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	IdleTimeout xtime.Duration
	// If WaitTimeout is set and the pool is at the Active limit, then
	// Get() waits WaitTimeout until a item to be returned to the pool
	// before returning.
	WaitTimeout xtime.Duration
	// If WaitTime not set, then Wait effects.
	// If Wait is set true, then wait until ctx timeout, or default false
	// and return directly.
	Wait bool
}

type item struct {
	createAt time.Time
	c        io.Closer
}

func (i *item) expired(timeout time.Duration) bool {
	if timeout < 0 {
		return false
	}
	return i.createAt.Add(timeout).Before(nowFunc())
}

func (i *item) close() error {
	return i.c.Close()
}

package breaker

import (
	xtime "github.com/zombie-k/kylin/library/time"
	"sync"
	"time"
)

const (
	// StateOpen when circuit breaker open, request not allowed, after sleep som duration,
	// allow one request for testing health, if ok then state reset to closed, if not
	// continue the step.
	StateOpen int32 = iota

	// StateClosed when circuit breaker closed, request allowed, the breaker calc the
	// succeed ratio, if request num greater than request setting and ratio lower than
	// the setting ratio, then reset state to open.
	StateClosed
)

type Config struct {
	// Breaker switch, default off.
	SwitchOff bool

	// Percentage of failures must lower than 1-1/K
	K float64

	Window  xtime.Duration
	Bucket  int
	Request int64
}

func (conf *Config) check() {
	if conf.K == 0 {
		conf.K = 1.5
	}
	if conf.Request == 0 {
		conf.Request = 100
	}
	if conf.Bucket == 0 {
		conf.Bucket = 10
	}
	if conf.Window == 0 {
		conf.Window = xtime.Duration(3 * time.Second)
	}
}

// Breaker is a CircuitBreaker pattern.
type Breaker interface {
	Allow() error
	MarkSuccess()
	MarkFailed()
}

// Group represents a class of CircuitBreaker and forms a namespace in which
// units of CircuitBreaker.
type Group struct {
	mu       sync.RWMutex
	breakers map[string]Breaker
	conf     *Config
}

var (
	_mu   sync.RWMutex
	_conf = &Config{
		K:       1.5,
		Window:  xtime.Duration(3 * time.Second),
		Bucket:  10,
		Request: 100,
	}
	_group = NewGroup(_conf)
)

func Init(conf *Config) {
	if conf == nil {
		return
	}
	_mu.Lock()
	_conf = conf
	_mu.Unlock()
}

func Go(name string, run, fallback func() error) error {
	breaker := _group.Get(name)
	if err := breaker.Allow(); err != nil {
		return fallback()
	}
	return run()
}

func newBreaker(c *Config) (b Breaker) {
	return newSRE(c)
}

func NewGroup(conf *Config) *Group {
	if conf == nil {
		_mu.RLock()
		conf = _conf
		_mu.RUnlock()
	} else {
		conf.check()
	}
	return &Group{
		conf:     conf,
		breakers: make(map[string]Breaker),
	}
}

func (g *Group) Get(key string) Breaker {
	g.mu.RLock()
	brk, ok := g.breakers[key]
	conf := g.conf
	g.mu.RUnlock()
	if ok {
		return brk
	}
	brk = newBreaker(conf)
	g.mu.Lock()
	if _, ok := g.breakers[key]; !ok {
		g.breakers[key] = brk
	}
	g.mu.Unlock()
	return brk
}

func (g *Group) Reload(conf *Config) {
	if conf == nil {
		return
	}
	conf.check()
	g.mu.Lock()
	g.conf = conf
	g.breakers = make(map[string]Breaker, len(g.breakers))
	g.mu.Unlock()
}

func (g *Group) Go(name string, run, fallback func() error) error {
	breaker := g.Get(name)
	if err := breaker.Allow(); err != nil {
		return fallback()
	}
	return run()
}

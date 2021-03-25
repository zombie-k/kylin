package breaker

import (
	"errors"
	"github.com/zombie-k/kylin/library/state/metric"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type sreBreaker struct {
	stat     metric.RollingCounter
	r        *rand.Rand
	randLock sync.Mutex

	k       float64
	request int64

	state int32
}

func (s *sreBreaker) Allow() error {
	success, total := s.summary()
	k := s.k * success
	if total < s.request || float64(total) < k {
		if atomic.LoadInt32(&s.state) == StateOpen {
			atomic.CompareAndSwapInt32(&s.state, StateOpen, StateClosed)
		}
		return nil
	}
	if atomic.LoadInt32(&s.state) == StateClosed {
		atomic.CompareAndSwapInt32(&s.state, StateClosed, StateOpen)
	}
	dropRatio := math.Max(0, (float64(total)-k)/float64(total+1))
	drop := s.dropOnRatio(dropRatio)
	if drop {
		return errors.New("ServiceUnavailable")
	}
	return nil
}

func (s *sreBreaker) MarkSuccess() {
	s.stat.Add(1)
}

func (s *sreBreaker) MarkFailed() {
	s.stat.Add(0)
}

func newSRE(c *Config) Breaker {
	counterOpts := metric.RollingCounterOpts{
		Size:           c.Bucket,
		BucketDuration: time.Duration(int64(c.Window) / int64(c.Bucket)),
	}
	stat := metric.NewRollingCounter(counterOpts)
	return &sreBreaker{
		stat:    stat,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
		k:       c.K,
		request: c.Request,
		state:   StateClosed,
	}
}

func (s *sreBreaker) summary() (success float64, total int64) {
	s.stat.Reduce(func(iterator metric.Iterator) float64 {
		for iterator.Next() {
			bucket := iterator.Bucket()
			total += bucket.Count
			for _, p := range bucket.Points {
				success += p
			}
		}
		return 0
	})
	return
}

func (s *sreBreaker) dropOnRatio(ratio float64) (b bool) {
	s.randLock.Lock()
	b = s.r.Float64() < ratio
	s.randLock.Unlock()
	return
}

package metric

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"sync/atomic"
)

// Counter stores a numerical value that only ever goes up.
type Counter interface {
	Metric
}

// CounterOpts is an alias of Opts.
type CounterOpts Opts

type counter struct {
	val int64
}

func (c *counter) Add(val int64) {
	if val < 0 {
		panic(fmt.Errorf("stat/metric: cannot decrease in negative value. val: %d", val))
	}
	atomic.AddInt64(&c.val, val)
}

func (c *counter) Value() int64 {
	return atomic.LoadInt64(&c.val)
}

func NewCounter(opts CounterOpts) Counter {
	return &counter{}
}

// CounterVecOpts is an alias of VectorOpts.
type CounterVecOpts VectorOpts

// CounterVec counter vec.
type CounterVec interface {
	// Incr increments the counter by 1. Use Add to increment it by arbitrary non-negative values.
	Incr(...string)
	// Add adds the given value to the counter. It panics if the value is < 0.
	Add(float64, ...string)
}

type prometheusCounterVec struct {
	Counter *prometheus.CounterVec
}

func (p *prometheusCounterVec) Incr(labels ...string) {
	p.Counter.WithLabelValues(labels...).Inc()
}

func (p *prometheusCounterVec) Add(val float64, labels ...string) {
	p.Counter.WithLabelValues(labels...).Add(val)
}

// NewCounterVec.
func NewCounterVec(cfg *CounterVecOpts) CounterVec {
	if cfg == nil {
		return nil
	}
	vec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: cfg.Namespace,
			Subsystem: cfg.SubSystem,
			Name:      cfg.Name,
			Help:      cfg.Help,
		},
		cfg.Labels)
	prometheus.MustRegister(vec)
	return &prometheusCounterVec{Counter: vec}
}

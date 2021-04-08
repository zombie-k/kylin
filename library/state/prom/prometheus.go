package prom

import "github.com/prometheus/client_golang/prometheus"

var (
	// LibClient for redis and db client
	LibClient = New().Namespace("go").Subsystem("lib_client").
		WithTimer("timer", []string{"method"}, nil).
		//WithSummary("summary", []string{"method"}).
		WithState("state", []string{"method", "name"}).
		WithCounter("code", []string{"method", "code"})

	//RPCClient rpc client
	RPCClient = New().Namespace("go").Subsystem("rpc_client").
		WithTimer("timer", []string{"method"}, nil).
		//WithSummary("summary", []string{"method"}).
		WithState("state", []string{"method", "name"}).
		WithCounter("code", []string{"method", "code"})

	//HTTPClient rpc client
	HTTPClient = New().Namespace("go").Subsystem("http_client").
		WithTimer("timer", []string{"method"}, nil).
		//WithSummary("summary", []string{"method"}).
		WithState("state", []string{"method", "name"}).
		WithCounter("code", []string{"method", "code"})
)

type Prom struct {
	namespace string
	subsystem string

	timer   *prometheus.HistogramVec
	counter *prometheus.CounterVec
	state   *prometheus.GaugeVec
	summary *prometheus.SummaryVec
}

// New creates a Prom instance.
func New() *Prom {
	return &Prom{}
}

func (p *Prom) Namespace(namespace string) *Prom {
	p.namespace = namespace
	return p
}
func (p *Prom) Subsystem(subsystem string) *Prom {
	p.subsystem = subsystem
	return p
}

func (p *Prom) WithTimer(name string, labels []string, bucket []float64) *Prom {
	if p == nil || p.timer != nil {
		return p
	}
	p.timer = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: p.namespace,
			Subsystem: p.subsystem,
			Name:      name,
			Help:      name,
			Buckets:   bucket,
		}, labels)
	prometheus.MustRegister(p.timer)
	return p
}

func (p *Prom) WithCounter(name string, labels []string) *Prom {
	if p == nil || p.counter != nil {
		return p
	}
	p.counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: p.namespace,
			Subsystem: p.subsystem,
			Name:      name,
			Help:      name,
		}, labels)
	prometheus.MustRegister(p.counter)
	return p
}

func (p *Prom) WithState(name string, labels []string) *Prom {
	if p == nil || p.state != nil {
		return p
	}
	p.state = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: p.namespace,
			Subsystem: p.subsystem,
			Name:      name,
			Help:      name,
		}, labels)
	prometheus.MustRegister(p.state)
	return p
}

func (p *Prom) WithSummary(name string, labels []string) *Prom {
	if p == nil || p.summary != nil {
		return p
	}
	p.summary = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  p.namespace,
		Subsystem:  p.subsystem,
		Name:       name,
		Help:       name,
		Objectives: map[float64]float64{0.9: 0.01, 0.99: 0.001, 0.999: 0.0001},
	}, labels)
	prometheus.MustRegister(p.summary)
	return p
}

func (p *Prom) Timing(name string, time int64, extra ...string) {
	label := append([]string{name}, extra...)
	if p.timer != nil {
		p.timer.WithLabelValues(label...).Observe(float64(time))
	}
}

func (p *Prom) Incr(name string, extra ...string) {
	label := append([]string{name}, extra...)
	if p.counter != nil {
		p.counter.WithLabelValues(label...).Inc()
	}
	if p.state != nil {
		p.state.WithLabelValues(label...).Inc()
	}
}

func (p *Prom) Decr(name string, extra ...string) {
	label := append([]string{name}, extra...)
	if p.state != nil {
		p.state.WithLabelValues(label...).Dec()
	}
}

func (p *Prom) State(name string, v int64, extra ...string) {
	label := append([]string{name}, extra...)
	if p.state != nil {
		p.state.WithLabelValues(label...).Set(float64(v))
	}
}

func (p *Prom) Add(name string, v int64, extra ...string) {
	label := append([]string{name}, extra...)
	if p.counter != nil {
		p.counter.WithLabelValues(label...).Add(float64(v))
	}
	if p.state != nil {
		p.state.WithLabelValues(label...).Add(float64(v))
	}
}
package metric

import "github.com/prometheus/client_golang/prometheus"

type HistogramVecOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
	Bucket    []float64
}

// HistogramVec gauge vec.
type HistogramVec interface {
	Observe(v int64, labels ...string)
}

// promHistogram prom histogram collection.
type promHistogramVec struct {
	histogram *prometheus.HistogramVec
}

func (p *promHistogramVec) Observe(v int64, labels ...string) {
	p.histogram.WithLabelValues(labels...).Observe(float64(v))
}

func NewHistogramVec(cfg *HistogramVecOpts) HistogramVec {
	if cfg == nil {
		return nil
	}
	vec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: cfg.Namespace,
			Subsystem: cfg.Subsystem,
			Name:      cfg.Name,
			Help:      cfg.Help,
			Buckets:   cfg.Bucket,
		}, cfg.Labels)
	prometheus.MustRegister(vec)
	return &promHistogramVec{histogram: vec}
}

package metric

import "github.com/prometheus/client_golang/prometheus"

// SummaryVec sets quantiles.
type SummaryVec interface {
	Observe(v int64, label ...string)
}

type promSummaryVec struct {
	summary *prometheus.SummaryVec
}

func (s *promSummaryVec) Observe(v int64, label ...string) {
	s.summary.WithLabelValues(label...).Observe(float64(v))
}

func NewSummaryVec(opts prometheus.SummaryOpts, labels []string) SummaryVec {
	sum := prometheus.NewSummaryVec(opts, labels)
	prometheus.MustRegister(sum)
	return &promSummaryVec{summary: sum}
}

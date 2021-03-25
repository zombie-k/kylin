package warden

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zombie-k/kylin/library/state/metric"
)

const (
	_clientNamespace = "http_client"
	_serverNamespace = "http_server"
)

var (
	_metricClientRequestQuantile = metric.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  _clientNamespace,
		Subsystem:  "requests",
		Name:       "quantile",
		Help:       "http client requests quantile",
		Objectives: map[float64]float64{0.5:0.05, 0.9:0.01, 0.99:0.001},
	}, []string{"path", "method"})

	_metricClientRequestDuration = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: _clientNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "http client requests duration(ms)",
		Labels:    []string{"path", "method"},
		Bucket:    []float64{5, 10, 25, 50, 100, 500, 1000},
	})
	_metricClientRequestCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: _clientNamespace,
		SubSystem: "requests",
		Name:      "code_total",
		Help:      "http client requests code count.",
		Labels:    []string{"path", "method", "code"},
	})
)
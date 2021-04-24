package metric

import "fmt"

// Opts contains the common arguments for creating Metric.
type Opts struct{}

// Metric is a sample interface.
// Implementations of Metrics in metric package are Counter, Gauge,
// PointGauge, RollingCounter, and RollingGauge.
type Metric interface {
	// Add adds the given value to then counter.
	Add(int64)

	// Value gets the current value.
	// If the metric's type is PointGauge, RollingCounter, RollingGauge,
	// It returns the sum value within the window.
	Value() int64
}

// Aggregation contains some common aggregation functions.
// Each aggregation can compute summary statistics of window.
type Aggregation interface {
	// Min finds the min value within the window.
	Min() float64
	// Max finds the max value within the window.
	Max() float64
	// Avg computes average value within the window.
	Avg() float64
	// Sum computes sum value within the window.
	Sum() float64
}

// VectorOpts contains the common arguments for creating vec Metric.
type VectorOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
}

const (
	_businessNamespace          = "Business"
	_businessSubSystemCount     = "Count"
	_businessSubSystemGauge     = "Gauge"
	_businessSubSystemHistogram = "Histogram"
)

func NewBusinessMetricCount(name string, labels ...string) CounterVec {
	if name == "" || len(labels) == 0 {
		panic("stat:BusinessMetricCount name should not be empty or labels length should be greater than zero")
	}
	return NewCounterVec(&CounterVecOpts{
		Namespace: _businessNamespace,
		Subsystem: _businessSubSystemCount,
		Name:      name,
		Labels:    labels,
		Help:      fmt.Sprintf("Business metric count %s", name),
	})
}

package metric

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	_ "github.com/prometheus/client_golang/prometheus/testutil"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestCounter(t *testing.T) {
	counter := NewCounter(CounterOpts{})
	count := rand.Intn(100)
	for i := 0; i < count; i++ {
		counter.Add(1)
	}
	val := counter.Value()
	assert.Equal(t, val, int64(count))
}

func TestCounterVec(t *testing.T) {
	counterVec := NewCounterVec(&CounterVecOpts{
		Namespace: "TestNamespace",
		SubSystem: "TestSubsystem",
		Name:      "TestName",
		Help:      "this is test metrics.",
		Labels:    []string{"name", "addr"},
	})
	counterVec.Incr("name1", "127.0.0.1")
	counterVec.Incr("name1", "127.0.0.1")
	counterVec.Incr("name1", "127.0.0.1")
	//c := counterVec.(*prometheusCounterVec)
	//ss , err := c.Counter.GetMetricWithLabelValues("name1", "127.0.0.1")
	//fmt.Printf("%s,%s\n", ss, err)
	//fmt.Printf("%+v,%+v\n", ss, err)
	mfs, err := prometheus.DefaultGatherer.Gather()
	fmt.Println(mfs, err)

	assert.Panics(t, func() {
		NewCounterVec(&CounterVecOpts{
			Namespace: "TestNamespace",
			SubSystem: "TestSubsystem",
			Name:      "TestName",
			Help:      "this is test metrics.",
			Labels:    []string{"name", "addr"},
		})
	}, "Expected to panic")

}

func TestPrometheus(t *testing.T) {
	v := NewCounterVec(&CounterVecOpts{
		Namespace: "TestNamespace",
		SubSystem: "TestSubsystem",
		Name:      "TestName",
		Help:      "this is test metrics.",
		Labels:    []string{"name", "addr"},
	})
	v.Incr("name1", "127.0.0.1")
	v.Incr("name2", "127.0.0.1")
	v.Incr("name3", "127.0.0.1")

	v1 := NewCounterVec(&CounterVecOpts{
		Namespace: "VideoRecommend",
		SubSystem: "827",
		Name:      "Test",
		Help:      "this is test metrics.",
		Labels:    []string{"name", "addr"},
	})
	v1.Incr("name2", "127.0.0.1")
	v1.Incr("name3", "127.0.0.1")
	v1.Incr("name1", "127.0.0.1")
	v1.Incr("name1", "127.0.0.1")
	v1.Incr("name4", "127.0.0.1")

	reg := prometheus.NewPedanticRegistry()
	reg.Register(v1.(*prometheusCounterVec).Counter)
	//got, _ := reg.Gather()
	got2, _ := prometheus.DefaultGatherer.Gather()
	for _, i := range got2 {
		if *i.Type != io_prometheus_client.MetricType_COUNTER {
			continue
		}
		fmt.Println(i)

		for _, vv := range i.Metric {
			fmt.Println(vv)
		}
	}

	c := testutil.CollectAndCount(v1.(*prometheusCounterVec).Counter, "VideoRecommend_827_Test")
	fmt.Println(c)
}

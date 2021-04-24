package stat

import (
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"testing"
	"time"
)

func TestStat(t *testing.T) {
	HTTPClient.Incr("/root/metrics", "200")
	HTTPClient.Incr("/root/metrics", "200")
	HTTPClient.Timing("/root/metrics", 55)
	time.Sleep(time.Millisecond * 20)
	HTTPClient.Timing("/root/metrics", 60)

	got2, _ := prometheus.DefaultGatherer.Gather()
	for _, i := range got2 {
		if strings.Contains(*i.Name, "go_memstats") ||
			*i.Name == "go_gc_duration_seconds" ||
			*i.Name == "go_goroutines" ||
			*i.Name == "go_info" ||
			*i.Name == "go_threads" {
			continue
		}
		/*
			if *i.Type != io_prometheus_client.MetricType_COUNTER {
				continue
			}
		*/
		t.Logf("%s", i)

		for _, vv := range i.Metric {
			t.Logf("%s", vv)
		}
	}
}

package warden

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	xtime "github.com/zombie-k/kylin/library/time"
	xhttp "net/http"
	"net/url"
	"os"
	"runtime"
	"testing"
	"time"
)

func defaultServer() {
	c := &ServerConfig{
		Network:      "tcp",
		Addr:         "0.0.0.0:19960",
		Timeout:      xtime.Duration(time.Second),
		ReadTimeout:  xtime.Duration(time.Second),
		WriteTimeout: xtime.Duration(time.Second),
	}
	engine := NewServer(c)
	engine.Start()
}

func TestT1(t *testing.T) {
	n, _ := os.Hostname()
	version := runtime.Version()
	t.Logf("n:%s, version:%s", n, version)
}

var (
	_conf = &ClientConfig{
		Dial:      xtime.Duration(time.Second),
		Timeout:   xtime.Duration(time.Second),
		KeepAlive: xtime.Duration(time.Second * 10),
	}
)

func TestClient(t *testing.T) {
	client := NewClient(_conf)
	uri := "http://10.182.9.194:6009/motan_client_demo"
	params := url.Values{}
	params.Set("uid", "5014976811")

	var res interface{}
	err := client.Get(context.Background(), uri, params, &res)
	t.Logf("err:%s", err)
	t.Logf("res:%+v", res)
	fmt.Println(res)
}

func TestClient1(t *testing.T) {
	defaultServer()
	client := NewClient(_conf)
	uri := "http://10.182.9.194:6009/motan_client_demo"
	params := url.Values{}
	params.Set("uid", "5014976811")
	req, err := client.NewRequest(xhttp.MethodGet, uri, params)
	if err != nil {
		t.Logf("err:%s", err)
		return
	}
	var bs []byte
	ticker := time.NewTicker(time.Second * 1)
	go func() {
		for {
			select {
			case <-ticker.C:
				b := time.Now()
				bs, err = client.Raw(context.Background(), req)
				cost := time.Since(b)
				fmt.Println("cost:", cost)
			}
		}
	}()
	time.Sleep(time.Second * 3)
	got2, _ := prometheus.DefaultGatherer.Gather()
	for _, i := range got2 {
		if *i.Type != io_prometheus_client.MetricType_COUNTER {
			continue
		}
		t.Logf("%s", i)
		for _, vv := range i.Metric {
			t.Logf("%s", vv)
		}
	}
	fmt.Println(bs)
	select {}
}

package warden

import (
	"fmt"
	"github.com/zombie-k/kylin/library/stat/metric"
	xtime "github.com/zombie-k/kylin/library/time"
	"net/http"
	"strings"
	"testing"
	"time"
)

type TypeName []struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

var (
	_metricClientReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "demo_api",
		Subsystem: "count",
		Name:      "test",
		Help:      "http client requests code count.",
		Labels:    []string{"path", "method", "code"},
	})
)

func goIncr() {
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-ticker.C:
			_metricClientReqCodeTotal.Incr("http://localhost:19960/test", "GET", "200")
		}
	}
}

func TestNewServer(t *testing.T) {
	c := &ServerConfig{
		Network:      "tcp",
		Addr:         "0.0.0.0:19960",
		Timeout:      xtime.Duration(time.Second),
		ReadTimeout:  xtime.Duration(time.Second),
		WriteTimeout: xtime.Duration(time.Second),
	}

	engine := NewServer(c)
	engine.addRoute("GET", "/test", func(c *Context) {
		value := []struct {
			Name string
			Age  int
			Sex  string
		}{
			{
				Name: "GBProgrammer",
				Age:  20,
				Sex:  "male",
			},
		}
		c.JSON(http.StatusOK, value, nil)
	})
	engine.addRoute("GET", "/testing", func(c *Context) {
		c.String(200, "nihao %s", "hhahah")
	})
	engine.addRoute("GET", "/test/go", func(c *Context) {
		var resp = TypeName{
			{Id: 1, Name: "AA"},
			{Id: 2, Name: "BB"},
		}
		c.JSON(200, resp, nil)
	})
	engine.Start()

	go goIncr()

	root := engine.trees.get("GET")
	res := ""
	scanNode(root, 1, &res)
	select {}
}

func TestCase(t *testing.T) {
	addrs := "1.1.1.1,2.2.2.2,3.3.3.3"
	port := "22"
	addrSlice := strings.Split(addrs, ",")
	for i, _ := range addrSlice {
		addrSlice[i] += ":" + port
	}
	fmt.Println(addrSlice)
}

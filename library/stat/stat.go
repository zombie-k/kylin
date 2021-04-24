package stat

import "github.com/zombie-k/kylin/library/stat/prom"

type Stat interface {
	Timing(name string, time int64, extra ...string)
	Incr(name string, extra ...string)
	State(name string, val int64, extra ...string)
}

var (
	HTTPClient Stat = prom.HTTPClient
	Cache      Stat = prom.LibClient
	RPCClient       = prom.RPCClient
)

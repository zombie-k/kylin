package warden

import (
	"github.com/pkg/errors"
	"github.com/zombie-k/kylin/library/log"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		now := time.Now()
		req := c.Request
		path := req.URL.Path
		params := req.Form
		var quota float64
		if deadline, ok := c.Context.Deadline(); ok {
			quota = time.Until(deadline).Seconds()
		}

		c.Next()

		err := c.Error
		cerr := errors.Cause(err)
		cost := time.Since(now)

		// TODO: server qps/cost metrics

		errmsg := ""
		if err != nil {
			errmsg = err.Error()
		}
		log.Accessv(c,
			log.KVString("method", req.Method),
			log.KVString("ip", ""),
			log.KVString("path", path),
			log.KVString("params", params.Encode()),
			log.KVString("err_cause", cerr.Error()),
			log.KVString("err", errmsg),
			log.KVFloat64("timeout_quota", quota),
			log.KVFloat64("cost", cost.Seconds()),
			log.KVString("source", "access-log"),
		)
	}
}

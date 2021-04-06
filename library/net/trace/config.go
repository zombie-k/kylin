package trace

import (
	xtime "github.com/zombie-k/kylin/library/time"
	"time"
)

type Config struct {
	// Report network eg: unixgram, tcp, udp
	Network string
	// For TCP and UDP networks, the addr has the form "host:port".
	// For Unix networks, the address must be a file system path.
	Addr string
	// Report timeout
	Timeout xtime.Duration
	// Close the sampling
	DisableSample bool
	// ProtocolVersion
	ProtocolVersion int32
	// Probability probability sampling
	Probability float32
}

var defaultOption = option{}

type option struct {
	Debug bool
}

type Option func(*option)

func EnableDebug() Option {
	return func(o *option) {
		o.Debug = true
	}
}

func Init(cfg *Config) {
	if cfg == nil {
		return
	}
	report := newReport(cfg.Network, cfg.Addr, time.Duration(cfg.Timeout), cfg.ProtocolVersion)
	SetGlobalTracer(NewTracer("VideoRecommend", report, cfg.DisableSample))
}

package log

import (
	"github.com/zombie-k/kylin/library/log/internel/filewriter"
)

var (
	logH Handler
)

func init() {
	logH = newHandlers([]string{}, NewStdout(""))
}

func Init(conf *Config) {
	if conf == nil {
		conf = &Config{
			Dir:          "",
			Stdout:       true,
			Rotate:       false,
			RotateFormat: "daily",
			Pattern:      "%T\t%L\t%M",
		}
	}

	var hs []Handler

	if conf.Stdout {
		hs = append(hs, NewStdout(conf.Pattern))
	}

	if conf.Dir != "" {
		if conf.Pattern == "" {
			conf.Pattern = "%T\t%L\t%M"
		}
		var options []filewriter.Option
		options = append(options, filewriter.SetRotate(conf.Rotate))
		if conf.RotateFormat == "hourly" {
			options = append(options, filewriter.SetRotateHourly(true))
		} else if conf.RotateFormat == "daily" {
			options = append(options, filewriter.SetRotateDaily(true))
		}
		hs = append(hs, NewFile(conf.Dir, conf.Pattern, options...))
	}

	if len(hs) > 0 {
		logH = newHandlers([]string{}, hs...)
	}
}

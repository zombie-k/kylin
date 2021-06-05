package log

import (
	"github.com/zombie-k/kylin/library/log/internel/filewriter"
)

var (
	logH Handler
	fH   FHandler
)

func init() {
	logH = newHandlers([]string{}, NewStdout("%M"))
}

func Init(conf *Config) {
	if conf == nil {
		conf = &Config{
			Dir:          "",
			Stdout:       true,
			Rotate:       false,
			RotateFormat: "daily",
			Pattern:      "%D\t%L\t%M",
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

func Init2(conf *Config) {
	if conf == nil {
		return
	}
	if conf.Pattern == "" {
		conf.Pattern = "%M"
	}
	var options []filewriter.Option
	options = append(options, filewriter.SetRotate(conf.Rotate))
	if conf.RotateFormat == "hourly" {
		options = append(options, filewriter.SetRotateHourly(true))
	} else if conf.RotateFormat == "daily" {
		options = append(options, filewriter.SetRotateDaily(true))
	}
	if conf.Suffix == "" {
		conf.Suffix = ".log"
	}
	cf := NewCustomFile(conf.Dir, conf.CustomFiles, conf.Suffix, conf.Pattern, options...)
	fH = newFHandlers(cf)
}

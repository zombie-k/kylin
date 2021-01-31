package filewriter

import "time"

const (
	RotateDaily    = "20060102"
	RotateHourly   = "2006010215"
	RotateMinutely = "200601021504"
	RotateSecondly = "20060102150405"
)

var defaultOption = option{
	Rotate:         true,
	RotateDaily:    true,
	RotateInterval: time.Second * 10,
	ChanSize:       1024 * 8,
}

type option struct {
	// Rotate logfile
	Rotate bool
	// Rotate daily
	RotateDaily bool
	// Rotate hourly
	RotateHourly bool
	// Rotate minutely
	RotateMinutely bool
	// Rotate interval time
	RotateInterval time.Duration

	// Write timeout
	WriteTimeout time.Duration
	// Channel size
	ChanSize int
}

type Option func(opt *option)

func SetRotate(rotate bool) Option {
	return func(opt *option) {
		opt.Rotate = rotate
	}
}

func SetRotateMinutely(minutely bool) Option {
	return func(opt *option) {
		opt.RotateMinutely = minutely
	}
}

func SetRotateHourly(hourly bool) Option {
	return func(opt *option) {
		opt.RotateHourly = hourly
	}
}

func SetRotateDaily(daily bool) Option {
	return func(opt *option) {
		opt.RotateDaily = daily
	}
}

func SetChanSize(chanSize int) Option {
	return func(opt *option) {
		opt.ChanSize = chanSize
	}
}

func SetRotateInterval(interval time.Duration) Option {
	return func(opt *option) {
		opt.RotateInterval = interval
	}
}

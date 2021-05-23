package parser

import (
	"time"
)

type Basic struct {
	Sleep time.Duration
}

type BasicOption struct {
	F func (*Basic)
}

func BasicSleepingTime(d time.Duration) BasicOption {
	return BasicOption{func(b *Basic) {
		b.Sleep = d
	}}
}
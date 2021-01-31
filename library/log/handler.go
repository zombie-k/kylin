package log

import (
	"context"
	"github.com/pkg/errors"
)

const (
	_level  = "level"
	_log    = "log"
	_time   = "time"
	_source = "source"
)

type Handler interface {
	Log(context.Context, Level, ...D)

	SetFormat(string)

	Close() error
}

type Handlers struct {
	filters  map[string]struct{}
	handlers []Handler
}

func newHandlers(filters []string, handlers ...Handler) *Handlers {
	filterSets := make(map[string]struct{})
	for _, k := range filters {
		filterSets[k] = struct{}{}
	}
	return &Handlers{filters: filterSets, handlers: handlers}
}

func (hs Handlers) Log(ctx context.Context, lv Level, d ...D) {
	d = append(d, KVString(_level, lv.String()))
	for _, h := range hs.handlers {
		h.Log(ctx, lv, d...)
	}
}

func (hs Handlers) Close() (err error) {
	for _, h := range hs.handlers {
		if e := h.Close(); e != nil {
			err = errors.WithStack(e)
		}
	}
	return
}

func (hs Handlers) SetFormat(format string) {
	for _, h := range hs.handlers {
		h.SetFormat(format)
	}
}

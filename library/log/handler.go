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

// FHandler CustomFile Handler interface.
type FHandler interface {
	File(context.Context, string, ...D)

	SetFormat(string)

	Close() error
}

type FHandlers struct {
	handler FHandler
}

func newFHandlers(handler FHandler) *FHandlers {
	return &FHandlers{handler: handler}
}

func (hs FHandlers) File(ctx context.Context, file string, d ...D) {
	hs.handler.File(ctx, file, d...)
}

func (hs FHandlers) Close() (err error) {
	if e := hs.handler.Close(); e != nil {
		err = errors.WithStack(e)
	}
	return
}

func (hs FHandlers) SetFormat(format string) {
	hs.handler.SetFormat(format)
}

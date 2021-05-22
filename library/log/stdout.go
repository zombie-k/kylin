package log

import (
	"context"
	"os"
)

var defaultPattern = "%D\t%L\t%M"

type StdoutHandler struct {
	render Render
}

func NewStdout(format string) *StdoutHandler {
	if format == "" {
		format = defaultPattern
	}
	return &StdoutHandler{render: newPatternRender(format)}
}

func (h *StdoutHandler) Log(ctx context.Context, lv Level, args ...D) {
	d := DToMap(args...)
	h.render.Render(os.Stderr, d)
}

func (h *StdoutHandler) Close() error {
	return nil
}

func (h *StdoutHandler) SetFormat(format string) {
	h.render = newPatternRender(format)
}

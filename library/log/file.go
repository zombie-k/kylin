package log

import (
	"context"
	"github.com/zombie-k/kylin/library/log/internel/filewriter"
	"io"
	"path/filepath"
)

const (
	_idxInfo = iota
	_idxWarn
	_idxError
	_idxMax
)

var _logFileNames = map[int]string{
	_idxInfo:  "info.log",
	_idxWarn:  "warn.log",
	_idxError: "error.log",
}

type FileHandler struct {
	render Render
	fws    []*filewriter.LogFileWriter

	// For CustomFile
	fwc map[string]*filewriter.LogFileWriter
}

func NewFile(dir string, pattern string, options ...filewriter.Option) *FileHandler {
	newWriter := func(name string) *filewriter.LogFileWriter {
		w, err := filewriter.NewLogFileWriter(filepath.Join(dir, name), options...)
		if err != nil {
			panic(err)
		}
		return w
	}

	handler := &FileHandler{
		fws:    make([]*filewriter.LogFileWriter, 3, 100),
		render: newPatternRender(pattern),
	}

	for idx, name := range _logFileNames {
		handler.fws[idx] = newWriter(name)
	}

	return handler
}

func NewCustomFile(dir string, customFiles []string, suffix, pattern string, options ...filewriter.Option) *FileHandler {
	newWriter := func(name string) *filewriter.LogFileWriter {
		w, err := filewriter.NewLogFileWriter(filepath.Join(dir, name), options...)
		if err != nil {
			panic(err)
		}
		return w
	}

	handler := &FileHandler{
		fwc:    make(map[string]*filewriter.LogFileWriter),
		render: newPatternRender(pattern),
	}

	for _, name := range customFiles {
		handler.fwc[name] = newWriter(name + suffix)
	}

	return handler
}

func (h *FileHandler) Log(ctx context.Context, lv Level, args ...D) {
	d := DToMap(args...)
	var w io.Writer
	switch lv {
	case _warnLevel:
		w = h.fws[_idxWarn]
	case _errorLevel:
		w = h.fws[_idxError]
	default:
		w = h.fws[_idxInfo]
	}

	h.render.Render(w, d)
}

func (h *FileHandler) File(ctx context.Context, file string, args ...D) {
	if w, ok := h.fwc[file]; ok {
		d := DToMap(args...)
		h.render.Render(w, d)
	}
}

func (h *FileHandler) Close() error {
	for _, h := range h.fws {
		h.Close()
	}
	for _, h := range h.fwc {
		h.Close()
	}
	return nil
}

func (h *FileHandler) SetFormat(format string) {
	h.render = newPatternRender(format)
}

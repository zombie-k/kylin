package log

import (
	"context"
	"fmt"
	"io"
)

type Config struct {
	Dir    string
	Stdout bool

	Rotate       bool
	RotateFormat string

	Pattern string
}

type Render interface {
	Render(w io.Writer, d map[string]interface{}) error
	RenderString(w io.Writer, d map[string]interface{}) string
}

func Info(format string, args ...interface{}) {
	logH.Log(context.Background(), _infoLevel, KVString(_log, fmt.Sprintf(format, args...)))
}

func Warn(format string, args ...interface{}) {
	logH.Log(context.Background(), _warnLevel, KVString(_log, fmt.Sprintf(format, args...)))
}

func Error(format string, args ...interface{}) {
	logH.Log(context.Background(), _errorLevel, KVString(_log, fmt.Sprintf(format, args...)))
}

func Infov(ctx context.Context, args ...D) {
	logH.Log(ctx, _infoLevel, args...)
}

func Warnv(ctx context.Context, args ...D) {
	logH.Log(ctx, _warnLevel, args...)
}

func Errorv(ctx context.Context, args ...D) {
	logH.Log(ctx, _errorLevel, args...)
}

func SetFormat(format string) {
	logH.SetFormat(format)
}

func Close() error {
	err := logH.Close()
	return err
}

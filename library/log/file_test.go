package log

import (
	"context"
	"testing"
	"time"
)

func TestFileHandler_Close(t *testing.T) {

}

func TestFileHandler_Log(t *testing.T) {
	h := NewFile("./log", "%T %L %f %M")
	var hs []Handler
	hs = append(hs, h)
	handlers := newHandlers([]string{}, hs...)
	d := []D{KVString(_log, "hello world!!!")}
	handlers.Log(context.Background(), _infoLevel, d...)
	handlers.Log(context.Background(), _errorLevel, d...)
	time.Sleep(time.Second)
}

func TestFileHandler_SetFormat(t *testing.T) {
}

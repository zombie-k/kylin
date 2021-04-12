package log

import (
	"context"
	"testing"
	"time"
)

func initStdout() {
	conf := &Config{Stdout:true}
	Init(conf)
}

func initFile() {
	conf := &Config{
		Dir:          "/Users/xiangqian5/github/kylin/log",
		Stdout:       true,
		Rotate:       false,
		RotateFormat: "daily",
		Pattern:      "%T\t%L\t%M",
	}
	Init(conf)
}

type TestLog struct {
	A string
	B int
	C string
	D string
}

func TestLog1(t *testing.T) {
	Init(nil)
	for i := 0; i < 1000; i++ {
		go Info("%s %s", "hello world", "second")
	}
	time.Sleep(time.Second * 2)
}

func testLog(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		Error("hello %s", "world")
		Errorv(context.Background(), KV("key", 222), KV("key2", "test"))
	})
	t.Run("Warn", func(t *testing.T) {
		Warn("hello %s", "world")
		Warnv(context.Background(), KV("key", 222), KV("key2", "test"))
	})
	t.Run("Info", func(t *testing.T) {
		Info("hello %s", "world")
		Infov(context.Background(), KV("key", 222), KV("key2", "test"))
	})
}

func TestFile(t *testing.T) {
	initFile()
	testLog(t)
	time.Sleep(time.Second*1)
}
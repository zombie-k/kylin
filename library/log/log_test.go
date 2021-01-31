package log

import (
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	Init(nil)
	for i := 0; i < 1000; i++ {
		go Info("%s %s", "hello world", "second")
	}
	time.Sleep(time.Second * 2)
}

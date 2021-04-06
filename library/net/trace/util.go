package trace

import (
	"context"
	"encoding/binary"
	"github.com/pkg/errors"
	"github.com/zombie-k/kylin/library/libutil/hash"
	"math/rand"
	"time"
)

var _hostHash byte

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	_hostHash = byte(hash.JenkinsOneAtTimeHash("test"))
}

func getID() uint64 {
	var b [8]byte
	binary.BigEndian.PutUint32(b[4:], uint32(time.Now().UnixNano())>>8)
	b[4] = _hostHash
	binary.BigEndian.PutUint32(b[:4], uint32(rand.Int31()))
	return binary.BigEndian.Uint64(b[:])
}

func extendTag() (tags []Tag) {
	tags = append(tags,
		TagString("ip", "127.0.0.1"),
	)
	return
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type ctxKey string

var _ctxKey ctxKey = "kylin/net/trace.trace"

func FromContext(ctx context.Context) (t Trace, ok bool) {
	t, ok = ctx.Value(_ctxKey).(Trace)
	return
}

func NewContext(ctx context.Context, t Trace) context.Context {
	return context.WithValue(ctx, _ctxKey, t)
}

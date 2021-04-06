package trace

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	flagSampled = 0x01
	flagDebug   = 0x02
)

var (
	errEmptyTracerString   = errors.New("trace: Can not convert empty string to SpanContext")
	errInvalidTracerString = errors.New("trace: String does not match SpanContext string format")
)

type SpanContext struct {
	// TraceID represents globally unique ID of the trace.
	// Usually generated as a random number.
	TraceID uint64

	// SpanID represents span ID that must be unique within its trace,
	// but does not have to be globally unique.
	SpanID uint64

	// ParentID refers to the ID of the parent span.
	// Should be 0 if the current span is a root span.
	ParentID uint64

	// Flags is a bitmap containing such bits as `sampled` and `capture`.
	Flags byte

	// Probability
	Probability float32

	// Level current level.
	Level int
}

func (sc SpanContext) isSampled() bool {
	return (sc.Flags & flagSampled) == flagSampled
}

func (sc SpanContext) isDebug() bool {
	return (sc.Flags & flagDebug) == flagDebug
}

func (sc SpanContext) IsValid() bool {
	return sc.TraceID != 0 && sc.SpanID != 0
}

var emptySpanContext = SpanContext{}

// String convert SpanContext to String
// {TraceID}:{SpanID}:{ParentID}:{flags}:[extend...]
// TraceID: uint64 base16
// SpanID: uint64 base16
// ParentID: uint64 base16
// flags:
// - :0 sampled flag
// - :1 capture flag
// extend
// sample-rate: s-{base16(BigEndian(float32))}
func (sc SpanContext) String() string {
	base := make([]string, 4)
	base[0] = strconv.FormatUint(sc.TraceID, 16)
	base[1] = strconv.FormatUint(sc.SpanID, 16)
	base[2] = strconv.FormatUint(sc.ParentID, 16)
	base[3] = strconv.FormatUint(uint64(sc.Flags), 16)
	return strings.Join(base, ":")
}

// ExtractContextFromString parse spanContext from string
func extractContextFromString(val string) (SpanContext, error) {
	if val == "" {
		return emptySpanContext, errEmptyTracerString
	}
	items := strings.Split(val, ":")
	if len(items) < 4 {
		return emptySpanContext, errInvalidTracerString
	}

	parseHexUint64 := func(hexs []string) ([]uint64, error) {
		ret := make([]uint64, len(hexs))
		var err error
		for i, hex := range hexs {
			ret[i], err = strconv.ParseUint(hex, 16, 64)
			if err != nil {
				break
			}
		}
		return ret, err
	}
	ret, err := parseHexUint64(items[0:4])
	if err != nil {
		return emptySpanContext, errInvalidTracerString
	}
	sctx := SpanContext{
		TraceID:  ret[0],
		SpanID:   ret[1],
		ParentID: ret[2],
		Flags:    byte(ret[3]),
	}
	return sctx, nil
}

func (sc *SpanContext) Format() string {
	return fmt.Sprintf("%d %d %d %d %d %f", sc.TraceID, sc.SpanID, sc.ParentID, sc.Level, sc.Flags, sc.Probability)
}

package trace

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	protogen "github.com/zombie-k/kylin/library/net/trace/proto"
	"math"
	"time"
)

const protoVersion int32 = 1

var errSpanVersion = errors.New("trace: marshal not support version")

func marshalSpan(sp *Span, version int32) ([]byte, error) {
	if version == protoVersion {
		return marshalSpanVersion(sp)
	}
	return nil, errSpanVersion
}

func marshalSpanVersion(sp *Span) ([]byte, error) {
	protoSpan := &protogen.Span{
		Version:       protoVersion,
		ServiceName:   sp.tracerAdaptor.serviceName,
		OperationName: sp.operationName,
		TraceId:       sp.context.TraceID,
		SpanId:        sp.context.SpanID,
		ParentId:      sp.context.ParentID,
		StartTime: &timestamp.Timestamp{
			Seconds: sp.startTime.Unix(),
			Nanos:   int32(sp.startTime.Nanosecond()),
		},
		Duration: &duration.Duration{
			Seconds: int64(sp.duration / time.Second),
			Nanos:   int32(sp.duration % time.Second),
		},
		Tags: make([]*protogen.Tag, len(sp.tags)),
		Logs: sp.logs,
	}

	for i := range sp.tags {
		protoSpan.Tags[i] = toProtoTag(sp.tags[i])
	}
	return proto.Marshal(protoSpan)
}

func toProtoTag(tag Tag) *protogen.Tag {
	pTag := &protogen.Tag{Key: tag.Key}
	switch value := tag.Value.(type) {
	case string:
		pTag.Kind = protogen.Kind_STRING
		pTag.Value = []byte(value)
	case int:
		pTag.Kind = protogen.Kind_INT
		pTag.Value = serializeInt64(int64(value))
	case int32:
		pTag.Kind = protogen.Kind_INT
		pTag.Value = serializeInt64(int64(value))
	case int64:
		pTag.Kind = protogen.Kind_INT
		pTag.Value = serializeInt64(value)
	case bool:
		pTag.Kind = protogen.Kind_BOOL
		pTag.Value = serializeBool(value)
	case float32:
		pTag.Kind = protogen.Kind_FLOAT
		pTag.Value = serializeFloat64(float64(value))
	case float64:
		pTag.Kind = protogen.Kind_FLOAT
		pTag.Value = serializeFloat64(value)
	default:
		pTag.Kind = protogen.Kind_STRING
		pTag.Value = []byte(fmt.Sprintf("%v", tag.Value))
	}
	return pTag
}

func serializeInt64(v int64) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(v))
	return data
}

func serializeFloat64(v float64) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, math.Float64bits(v))
	return data
}

func serializeBool(v bool) []byte {
	data := make([]byte, 1)
	if v {
		data[0] = byte(1)
	} else {
		data[0] = byte(0)
	}
	return data
}

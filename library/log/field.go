package log

import (
	"github.com/zombie-k/kylin/library/log/internel/core"
	"github.com/zombie-k/kylin/library/time"
	"math"
)

type D = core.Field

func KVString(key string, value string) D {
	return D{Key: key, Type: core.StringType, StringVal: value}
}

func KVInt(key string, value int) D {
	return D{Key: key, Type: core.IntType, Int64Val: int64(value)}
}

func KVInt64(key string, value int64) D {
	return D{Key: key, Type: core.Int64Type, Int64Val: int64(value)}
}

func KVUint(key string, value uint) D {
	return D{Key: key, Type: core.UintType, Int64Val: int64(value)}
}

func KVUint64(key string, value uint64) D {
	return D{Key: key, Type: core.Uint64Type, Int64Val: int64(value)}
}

func KVFloat32(key string, value float32) D {
	return D{Key: key, Type: core.Float32Type, Int64Val: int64(math.Float32bits(value))}
}

func KVFloat64(key string, value float64) D {
	return D{Key: key, Type: core.Float64Type, Int64Val: int64(math.Float64bits(value))}
}

func KVDuration(key string, value time.Duration) D {
	return D{Key: key, Type: core.DurationType, Int64Val: int64(value)}
}

func KV(key string, value interface{}) D {
	return D{Key: key, Value: value}
}

func DToMap(args ...D) map[string]interface{} {
	d := make(map[string]interface{}, 10+len(args))
	for _, arg := range args {
		switch arg.Type {
		case core.UintType, core.Uint64Type, core.IntType, core.Int64Type:
			d[arg.Key] = arg.Int64Val
		case core.Float32Type:
			d[arg.Key] = math.Float32frombits(uint32(arg.Int64Val))
		case core.Float64Type:
			d[arg.Key] = math.Float64frombits(uint64(arg.Int64Val))
		case core.StringType:
			d[arg.Key] = arg.StringVal
		case core.DurationType:
			d[arg.Key] = time.Duration(arg.Int64Val)
		default:
			d[arg.Key] = arg.Value
		}
	}
	return d
}

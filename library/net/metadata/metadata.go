package metadata

import (
	"context"
	"fmt"
)

// MD is a mapping metadata keys to values
type MD map[string]interface{}

type mdKey struct{}

func (md MD) Len() int {
	return len(md)
}

func (md MD) Copy() MD {
	return Join(md)
}

func New(m map[string]interface{}) MD {
	md := MD{}
	for k, v := range m {
		md[k] = v
	}
	return md
}

func Join(mds ...MD) MD {
	out := MD{}
	for _, md := range mds {
		for k, v := range md {
			out[k] = v
		}
	}
	return out
}

// Pairs returns an MD formed by the mapping of key, value ...
// Pairs panics if len(kv) is odd.
func Pairs(kv ...interface{}) MD {
	if len(kv)%2 == 1 {
		panic(fmt.Sprintf("metadata: Pairs got the odd number of input pairs for metadata: %d", len(kv)))
	}
	md := MD{}
	var key string
	for k, v := range kv {
		if k%2 == 0 {
			key = v.(string)
			continue
		}
		md[key] = v
	}
	return md
}

// Create a new context with md attached.
func NewContext(ctx context.Context, md MD) context.Context {
	return context.WithValue(ctx, mdKey{}, md)
}

// return the incoming metadata in ctx if it exists. The
// returned MD should not be modified. Writing to it may cause races.
// Modification should be made to copies of the returned MD.
func FromContext(ctx context.Context) (md MD, ok bool) {
	md, ok = ctx.Value(mdKey{}).(MD)
	return
}

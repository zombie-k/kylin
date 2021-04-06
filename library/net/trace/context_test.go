package trace

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSpanContext(t *testing.T) {
	pctx := SpanContext{
		TraceID:  getID(),
		SpanID:   getID(),
		ParentID: getID(),
		Flags:    flagSampled,
	}
	t.Logf("pctx:%s", pctx)
	assert.Equal(t, true, pctx.isSampled())
	value := pctx.String()
	t.Logf("pctx:%s", value)
	pctx2, err := extractContextFromString(value)
	assert.NoError(t, err)
	t.Logf("pctx2:%s", pctx2)
	assert.Equal(t, pctx.TraceID, pctx2.TraceID)
	assert.Equal(t, pctx.SpanID, pctx2.SpanID)
	assert.Equal(t, pctx.ParentID, pctx2.ParentID)
	assert.Equal(t, pctx.Flags, pctx2.Flags)
}

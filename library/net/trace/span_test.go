package trace

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockReport struct {
	sps []*Span
}

func (m *mockReport) WriteSpan(sp *Span) error {
	m.sps = append(m.sps, sp)
	return nil
}

func (m *mockReport) Close() error {
	return nil
}

func TestSpan(t *testing.T) {
	report := &mockReport{}
	t1 := NewTracer("srv1", report, true)
	t.Run("test span string", func(t *testing.T) {
		sp1 := t1.New("opt1").(*Span)
		t.Logf("sp1:%s", sp1)
		fmt.Println(sp1.context.TraceID, sp1.context.SpanID, sp1.context.ParentID, sp1.context.Flags, sp1.context.Level)
	})
	t.Run("test fork", func(t *testing.T) {
		sp1 := t1.New("testfork").(*Span)
		sp2 := sp1.Fork("fork", "opt2").(*Span)
		fmt.Println(sp1.context.Format())
		fmt.Println(sp2.context.Format())
		assert.Equal(t, sp1.context.TraceID, sp2.context.TraceID)
		assert.Equal(t, sp1.context.SpanID, sp2.context.ParentID)
		t.Run("test max fork", func(t *testing.T) {
			sp3 := sp2.Fork("xxx", "xxxx")
			for i := 0; i < 100; i++ {
				sp3 = sp3.Fork("", "xxxxx")
			}
			assert.Equal(t, noopspan{}, sp3)
		})
		t.Run("test max childs", func(t *testing.T) {
			sp3 := sp2.Fork("xxx", "xxxx")
			for i := 0; i < 4096; i++ {
				sp3.Fork("", "xxx")
			}
			assert.Equal(t, noopspan{}, sp3.Fork("xx", "xx"))
		})
	})
}

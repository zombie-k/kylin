package trace

import "testing"

func TestMarshalSpan(t *testing.T) {
	report := &mockReport{}
	t1 := NewTracer("srv1", report, true)
	sp1 := t1.New("op1").(*Span)
	sp1.SetLog(Log("hello", "test123"))
	sp1.SetTag(TagString("tag1", "hell"), TagBool("booltag", true), TagFloat64("float64tag", 3.14159))
	sp1.Finish(nil)
	t.Logf("tags:%s", sp1.tags)
	t.Logf("logs:%s", sp1.logs)
	t.Logf("%s", sp1.startTime)
	t.Logf("%s", sp1.duration)
	t.Logf("%s", sp1.operationName)
	t.Logf("%v", sp1.childs)
	v, err := marshalSpanVersion(sp1)
	t.Logf("v:%s", v)
	t.Logf("e:%v", err)
}

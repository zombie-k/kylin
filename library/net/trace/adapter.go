package trace

import (
	"log"
	"os"
	"sync"
	"time"
)

const (
	_maxLevel    = 64
	_probability = 0.00025
)

type tracerAdaptor struct {
	serviceName   string
	disableSample bool
	tags          []Tag
	reporter      reporter
	propagators   map[interface{}]propagator
	pool          *sync.Pool
	stdLog        *log.Logger
	sampler       sampler
}

func NewTracer(serviceName string, report reporter, disableSample bool) Tracer {
	sampler := newSampler(_probability)

	// default internal tags
	tags := extendTag()
	stdLog := log.New(os.Stderr, "trace", log.LstdFlags)
	return &tracerAdaptor{
		serviceName:   serviceName,
		disableSample: disableSample,
		tags:          tags,
		reporter:      report,
		propagators: map[interface{}]propagator{
			HTTPFormat: httpPropagator{},
		},
		pool: &sync.Pool{New: func() interface{} {
			return new(Span)
		}},
		stdLog:  stdLog,
		sampler: sampler,
	}
}

func (ta *tracerAdaptor) New(operationName string, opts ...Option) Trace {
	var (
		traceID     uint64
		sampled     bool
		probability float32
	)

	opt := defaultOption
	for _, fn := range opts {
		fn(&opt)
	}
	traceID = getID()
	if ta.disableSample {
		sampled = true
		probability = 1
	} else {
		sampled, probability = ta.sampler.IsSampled(traceID, operationName)
	}
	spanCtx := SpanContext{TraceID: traceID}
	if sampled {
		spanCtx.Flags = flagSampled
		spanCtx.Probability = probability
	}
	if opt.Debug {
		spanCtx.Flags |= flagDebug
		return ta.newSpanWithContext(operationName, spanCtx)
	}
	return ta.newSpanWithContext(operationName, spanCtx)
}

func (ta *tracerAdaptor) newSpanWithContext(operationName string, pctx SpanContext) Trace {
	sp := ta.getSpan()
	if pctx.Level > _maxLevel {
		return noopspan{}
	}
	level := pctx.Level + 1
	nctx := SpanContext{
		TraceID:  pctx.TraceID,
		ParentID: pctx.SpanID,
		Flags:    pctx.Flags,
		Level:    level,
	}
	if pctx.SpanID == 0 {
		nctx.SpanID = pctx.TraceID
	} else {
		nctx.SpanID = getID()
	}
	sp.operationName = operationName
	sp.context = nctx
	sp.startTime = time.Now()
	sp.tags = append(sp.tags, ta.tags...)
	return sp
}

func (ta *tracerAdaptor) Inject(t Trace, format interface{}, carrier interface{}) error {
	// if carrier implement Carrier use direct, ignore format
	carr, ok := carrier.(Carrier)
	if ok {
		t.Visit(carr.Set)
		return nil
	}

	// use Built-in propagators
	pp, ok := ta.propagators[format]
	if !ok {
		return ErrUnsupportedFormat
	}
	carr, err := pp.Inject(carrier)
	if err != nil {
		return err
	}
	if t != nil {
		t.Visit(carr.Set)
	}
	return nil
}

func (ta *tracerAdaptor) Extract(format interface{}, carrier interface{}) (Trace, error) {
	sp, err := ta.extract(format, carrier)
	if err != nil {
		return sp, err
	}
	return sp.SetTag(TagString(TagSpanKind, "server")), nil
}

func (ta *tracerAdaptor) extract(format interface{}, carrier interface{}) (Trace, error) {
	// if carrier implement Carrier use direct, ignore format
	carr, ok := carrier.(Carrier)
	if !ok {
		// use Built-in propagators
		pp, ok := ta.propagators[format]
		if !ok {
			return nil, ErrUnsupportedFormat
		}
		var err error
		if carr, err = pp.Extract(carrier); err != nil {
			return nil, err
		}
	}
	pctx, err := extractContextFromString(carr.Get(KyLinTraceID))
	if err != nil {
		return nil, err
	}
	return ta.newSpanWithContext("", pctx), nil
}

func (ta *tracerAdaptor) Close() error {
	return ta.reporter.Close()
}

func (ta *tracerAdaptor) report(sp *Span) {
	if sp.context.isSampled() {
		if err := ta.reporter.WriteSpan(sp); err != nil {
			ta.stdLog.Printf("marshal trace span error: %s", err)
		}
	}
	ta.putSpan(sp)
}

func (ta *tracerAdaptor) putSpan(sp *Span) {
	if len(sp.tags) > 32 {
		sp.tags = nil
	}
	if len(sp.logs) > 32 {
		sp.logs = nil
	}
	ta.pool.Put(sp)
}

func (ta *tracerAdaptor) getSpan() *Span {
	sp := ta.pool.Get().(*Span)
	sp.tracerAdaptor = ta
	sp.childs = 0
	sp.tags = sp.tags[:0]
	sp.logs = sp.logs[:0]
	return sp
}

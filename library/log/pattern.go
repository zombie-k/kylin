package log

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

var patternMap = map[string]func(map[string]interface{}) string{
	"T": longTime,
	"t": shortTime,
	"D": longDate,
	"d": shortDate,
	//"s": shortSource,
	"L": keyFactory(_level),
	"f": keyFactory(_source),
	"M": message,
}

type pattern struct {
	funcFactory []func(map[string]interface{}) string
	bufPool     sync.Pool
}

func newPatternRender(format string) Render {
	p := &pattern{
		bufPool: sync.Pool{New: func() interface{} {
			return &strings.Builder{}
		}},
	}

	b := make([]byte, 0, len(format))
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			b = append(b, format[i])
			continue
		}
		if i+1 >= len(format) {
			b = append(b, format[i])
			continue
		}
		f, ok := patternMap[string(format[i+1])]
		if !ok {
			b = append(b, format[i])
			continue
		}
		if len(b) != 0 {
			p.funcFactory = append(p.funcFactory, factoryText(string(b)))
			b = b[:0]
		}

		p.funcFactory = append(p.funcFactory, f)
		i++
	}

	if len(b) != 0 {
		p.funcFactory = append(p.funcFactory, factoryText(string(b)))
	}
	return p
}

func (p *pattern) Render(w io.Writer, d map[string]interface{}) error {
	builder := p.bufPool.Get().(*strings.Builder)
	defer func() {
		builder.Reset()
		p.bufPool.Put(builder)
	}()

	for _, f := range p.funcFactory {
		builder.WriteString(f(d))
	}

	builder.Write([]byte("\n"))
	_, err := w.Write([]byte(builder.String()))
	return err
}

func (p *pattern) RenderString(w io.Writer, d map[string]interface{}) string {
	builder := p.bufPool.Get().(*strings.Builder)
	defer func() {
		builder.Reset()
		p.bufPool.Put(builder)
	}()

	for _, f := range p.funcFactory {
		builder.WriteString(f(d))
	}

	return builder.String()
}

func factoryText(text string) func(map[string]interface{}) string {
	return func(map[string]interface{}) string {
		return text
	}
}

func keyFactory(key string) func(map[string]interface{}) string {
	return func(d map[string]interface{}) string {
		if v, ok := d[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
			return fmt.Sprint(v)
		}
		return ""
	}
}

func longTime(map[string]interface{}) string {
	return time.Now().Format("15:04:05.000")
}

func shortTime(map[string]interface{}) string {
	return time.Now().Format("15:04:05")
}

func longDate(map[string]interface{}) string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}

func shortDate(map[string]interface{}) string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func IsInternalKey(key string) bool {
	switch key {
	case _time, _level, _source:
		return true
	}

	return false
}

func message(d map[string]interface{}) string {
	var m string
	var s []string
	for k, v := range d {
		if k == _log {
			m = fmt.Sprint(v)
			continue
		}
		if IsInternalKey(k) {
			continue
		}
		s = append(s, fmt.Sprintf("%s=%v", k, v))
	}
	s = append(s, m)

	return strings.Join(s, " ")
}

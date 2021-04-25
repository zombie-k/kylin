package zlib

import (
	"bytes"
	"compress/zlib"
	"io"
	"io/ioutil"
	"sync"
)

type Pool struct {
	reader *sync.Pool
	writer *sync.Pool
	buffer *sync.Pool
}

func New() *Pool {
	pool := &Pool{
		reader: &sync.Pool{
			New: func() interface{} { return bytes.NewReader([]byte{0}) }},
		writer: &sync.Pool{
			New: func() interface{} {
				writer, _ := zlib.NewWriterLevel(new(bytes.Buffer), zlib.DefaultCompression)
				return writer
			}},
		buffer: &sync.Pool{
			New: func() interface{} { return new(bytes.Buffer) },
		},
	}
	return pool
}

func (p *Pool) Compress(val []byte) ([]byte, error) {
	in := p.buffer.Get().(*bytes.Buffer)
	in.Reset()
	defer p.buffer.Put(in)

	w := p.writer.Get().(*zlib.Writer)
	w.Reset(in)
	w.Write(val)
	defer func() {
		w.Close()
		p.writer.Put(w)
	}()

	buf, err := ioutil.ReadAll(in)
	return buf, err
}

func (p *Pool) UnCompress(val []byte) ([]byte, error) {
	out := p.buffer.Get().(*bytes.Buffer)
	out.Reset()
	defer p.buffer.Put(out)

	r := p.reader.Get().(*bytes.Reader)
	r.Reset(val)

	zr, _ := zlib.NewReader(r)
	defer zr.Close()
	io.Copy(out, zr)

	buf, err := ioutil.ReadAll(out)
	return buf, err
}


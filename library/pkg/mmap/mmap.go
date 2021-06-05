// Package mmap allows mapping files into memory. It tries to provide a simple, reasonably portable interface,
// but doesn't go out of its way to abstract away every little platform detail.
// This specifically means:
//	* forked processes may or may not inherit mappings
//	* a file's timestamp may or may not be updated by writes through mappings
//	* specifying a size larger than the file's actual size can increase the file's size
//	* If the mapped file is being modified by another process while your program's running, don't expect consistent results between platforms
package mmap

import (
	"bytes"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"path"
	"sync"
)

const (
	RDONLY = 0
	RDWR   = 1 << iota
	COPY
	EXEC
)

type Processor interface {
	ProcessLine(line []byte) error
}

type MMap struct {
	c *Config

	mmap []byte
	f    *os.File

	pool *sync.Pool
	wg   *sync.WaitGroup
	ch   chan struct{}
}

// Map maps an entire file into memory.
// The offset parameter must be a multiple of the system's page size.
// If length < 0, the entire file will be mapped.
func Mmap(c *Config) (*MMap, error) {
	flag := os.O_RDONLY
	if c.File.Mode != "r" {
		flag |= os.O_RDWR | os.O_CREATE | os.O_APPEND
	}
	f, err := os.OpenFile(path.Join(c.File.Path, c.File.Name), flag, os.FileMode(c.File.Perm))
	if err != nil {
		return nil, err
	}
	length := c.Map.Length
	if c.Map.Length < 0 {
		fi, err := f.Stat()
		if err != nil {
			return nil, err
		}
		length = int(fi.Size())
	}

	b, err := mmap(length, c.Map.Prot, c.Map.Flags, f.Fd(), c.Map.Offset)
	if err != nil {
		return nil, err
	}
	m := &MMap{
		c:    c,
		mmap: b,
		f:    f,
		pool: &sync.Pool{New: func() interface{} {
			buf := make([]byte, c.Kernel.ReadBytes)
			return buf
		}},
		ch: make(chan struct{}, c.Kernel.Routines),
		wg: &sync.WaitGroup{},
	}
	return m, nil
}

func (m *MMap) Process(processor Processor) error {
	buf := bytes.NewBuffer(m.mmap)
	for {
		chunk := m.pool.Get().([]byte)
		n := copy(chunk, buf.Next(m.c.Kernel.ReadBytes))
		chunk = chunk[:n]
		nextUntilNewline, err := buf.ReadBytes('\n')
		if err != io.EOF {
			chunk = append(chunk, nextUntilNewline...)
		} else {
			if len(chunk) == 0 {
				break
			}
		}
		chunk = chunk[:len(chunk)-1]
		m.ch <- struct{}{}
		m.wg.Add(1)
		go m.processChunk(chunk, processor)
	}
	m.wg.Wait()
	return nil
}

func (m *MMap) processChunk(chunk []byte, processor Processor) {
	defer func() {
		m.pool.Put(chunk)
		m.wg.Done()
		<-m.ch
	}()
	linesSlice := bytes.Split(chunk, []byte{'\n'})
	for _, line := range linesSlice {
		processor.ProcessLine(line)
	}
}

func (m *MMap) Flush() error {
	return unix.Msync(m.mmap, unix.MS_SYNC)
}

func (m *MMap) Munmap() error {
	err := unix.Munmap(m.mmap)
	m.mmap = nil
	m.f.Close()
	close(m.ch)
	return err
}

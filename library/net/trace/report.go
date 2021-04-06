package trace

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

const (
	_maxPackageSize             = 1024 * 32
	_defaultChanSize            = 4096
	_defaultWriteChannelTimeout = 50 * time.Millisecond
	_defaultWriteTimeout        = 200 * time.Millisecond
)

// reporter trace reporter
type reporter interface {
	WriteSpan(sp *Span) error
	Close() error
}

type connReport struct {
	version int32
	mutex   sync.RWMutex
	closed  bool

	network string
	address string

	conn net.Conn

	dataChan chan []byte
	done     chan struct{}
	timeout  time.Duration
}

func (c *connReport) Close() error {
	c.mutex.Lock()
	c.closed = true
	c.mutex.Unlock()

	timer := time.NewTicker(time.Second)
	close(c.dataChan)
	select {
	case <-timer.C:
		_ = c.closeConn()
		return fmt.Errorf("close report timeout force close")
	case <-c.done:
		return c.closeConn()
	}
}

func newReport(network, address string, timeout time.Duration, protocolVersion int32) reporter {
	if timeout <= 0 {
		timeout = _defaultWriteTimeout
	}
	report := &connReport{
		version:  protocolVersion,
		network:  network,
		address:  address,
		dataChan: make(chan []byte, _defaultChanSize),
		done:     make(chan struct{}),
		timeout:  timeout,
	}
	go report.daemon()
	return report
}

func (c *connReport) daemon() {
	for b := range c.dataChan {
		c.send(b)
	}
	c.done <- struct{}{}
}

func (c *connReport) send(data []byte) {
	if c.conn == nil {
		if err := c.reconnect(); err != nil {
			c.Errorf("connect error: %s retry after second", err)
			time.Sleep(time.Second)
			return
		}
	}
	_ = c.conn.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
	if _, err := c.conn.Write(data); err != nil {
		c.Errorf("write to conn error: %s, close connect", err)
		c.conn.Close()
		c.conn = nil
	}
}

func (c *connReport) WriteSpan(sp *Span) error {
	data, err := marshalSpan(sp, c.version)
	if err != nil {
		return err
	}
	return c.writePackage(data)
}

func (c *connReport) writePackage(data []byte) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.closed {
		return fmt.Errorf("report already closed")
	}
	if len(data) > _maxPackageSize {
		return fmt.Errorf("package too large length %d > %d", len(data), _maxPackageSize)
	}
	select {
	case c.dataChan <- data:
		return nil
	case <-time.After(_defaultWriteChannelTimeout):
		return fmt.Errorf("write to data channel timeout")
	}
}

func (c *connReport) reconnect() (err error) {
	c.conn, err = net.DialTimeout(c.network, c.address, c.timeout)
	return
}

func (c *connReport) closeConn() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *connReport) Errorf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

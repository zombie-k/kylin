package filewriter

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type LogFileWriter struct {
	opt    option
	ch     chan *bytes.Buffer
	pool   *sync.Pool
	stdLog *log.Logger

	dir         string
	filename    string
	fileHandler *fdHandler

	RotateFormat     string
	lastRotateFormat string

	closed int32
	wg     sync.WaitGroup
}

type fdHandler struct {
	fd    *os.File
	fSize int64
}

func (fdh *fdHandler) Size() int64 {
	if fdh != nil {
		return fdh.fSize
	}
	return 0
}

func (fdh *fdHandler) Write(b []byte) (n int, err error) {
	n, err = fdh.fd.Write(b)
	fdh.fSize += int64(n)
	return
}

func (fdh *fdHandler) Close() error {
	if fdh != nil {
		if fdh.fd != nil {
			err := fdh.fd.Close()
			return err
		}
	}
	return nil
}

func NewFdHandler(fpath string) (*fdHandler, error) {
	fd, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	fi, err := fd.Stat()
	if err != nil {
		return nil, err
	}

	return &fdHandler{
		fd:    fd,
		fSize: fi.Size(),
	}, nil
}

func NewLogFileWriter(fpath string, fns ...Option) (*LogFileWriter, error) {
	opt := defaultOption
	for _, fn := range fns {
		fn(&opt)
	}

	fname := filepath.Base(fpath)
	if fname == "" {
		return nil, fmt.Errorf("filename can't empty")
	}
	dir := filepath.Dir(fpath)
	fi, err := os.Stat(dir)
	if err == nil && !fi.IsDir() {
		return nil, fmt.Errorf("%s already exists and not a directory", dir)
	}
	if os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create dir %s error: %s", dir, err.Error())
		}
	}

	w := &LogFileWriter{
		opt:      opt,
		pool:     &sync.Pool{New: func() interface{} { return new(bytes.Buffer) }},
		ch:       make(chan *bytes.Buffer, opt.ChanSize),
		dir:      dir,
		stdLog:   log.New(os.Stderr, "fwlog", log.LstdFlags),
		filename: fname,
		closed:   0,
	}

	if w.opt.RotateMinutely {
		w.RotateFormat = RotateMinutely
	} else if w.opt.RotateHourly {
		w.RotateFormat = RotateHourly
	} else if w.opt.RotateDaily {
		w.RotateFormat = RotateDaily
	}

	err = w.initRotate()
	w.wg.Add(1)
	go w.daemon()

	return w, err
}

func (w *LogFileWriter) Write(b []byte) (int, error) {
	if atomic.LoadInt32(&w.closed) == 1 {
		//log
		return 0, fmt.Errorf("LogFileWriter already closed")
	}

	buf := w.getBuf()
	buf.Write(b)

	if w.opt.WriteTimeout == 0 {
		select {
		case w.ch <- buf:
			return len(b), nil
		default:
			return 0, fmt.Errorf("log channel is full, discard log")
		}
	}

	timeout := time.NewTimer(time.Second)
	select {
	case w.ch <- buf:
		return len(b), nil
	case <-timeout.C:
		return 0, fmt.Errorf("write log timeout, discard log")
	}
}

func (w *LogFileWriter) Close() error {
	atomic.StoreInt32(&w.closed, 1)
	close(w.ch)
	w.fileHandler.Close()
	w.wg.Wait()
	return nil
}

func (w *LogFileWriter) daemon() {
	defer func() {
		w.wg.Done()
	}()
	//rotate ticker
	tk := time.NewTicker(time.Millisecond * 10)
	//aggregate data buffer
	aggregateBuf := &bytes.Buffer{}
	//aggregate data ticker
	aggregateTicker := time.NewTicker(time.Millisecond * 20)

	//rotate and write
	for {
		select {
		case t := <-tk.C:
			w.checkRotate(t)
		case <-aggregateTicker.C:
			//write aggregate buffer data
			if aggregateBuf.Len() > 0 {
				if _, err := w.fileHandler.Write(aggregateBuf.Bytes()); err != nil {
					w.stdLog.Printf("write log error: %s", err)
				}
				aggregateBuf.Reset()
			}
		case buf, ok := <-w.ch:
			if ok {
				aggregateBuf.Write(buf.Bytes())
				w.putBuf(buf)
			}
		}
		if atomic.LoadInt32(&w.closed) != 1 {
			continue
		}
		if _, err := w.fileHandler.Write(aggregateBuf.Bytes()); err != nil {
			w.stdLog.Printf("write log error: %s", err)
		}
		for buf := range w.ch {
			if _, err := w.fileHandler.Write(buf.Bytes()); err != nil {
				w.stdLog.Printf("write log error: %s", err)
			}
			w.putBuf(buf)
		}
		break
	}
}

func (w *LogFileWriter) initRotate() error {
	fpath := filepath.Join(w.dir, w.filename)
	fi, err := os.Stat(fpath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	now := time.Now()
	format := now.Format(w.RotateFormat)
	if w.opt.Rotate && !os.IsNotExist(err) {
		fname := ""

		if w.fileHandler.Size() > 0 {
			if format != w.lastRotateFormat && format != fi.ModTime().Format(w.RotateFormat) {
				if w.opt.RotateHourly {
					fname = fpath + fmt.Sprintf(".%s", now.Add(-time.Hour).Format(w.RotateFormat))
				} else if w.opt.RotateDaily {
					fname = fpath + fmt.Sprintf(".%s", now.AddDate(0, 0, -1).Format(w.RotateFormat))
				}
			}
		}

		if fname != "" {
			w.fileHandler.Close()
			if err := os.Rename(fpath, fname); err != nil {
				return fmt.Errorf("Rotate: %s\n", err)
			}
		}
	}

	fdh, err := NewFdHandler(fpath)
	if err != nil {
		return err
	}

	w.fileHandler = fdh
	w.lastRotateFormat = format

	return nil
}

func (w *LogFileWriter) checkRotate(t time.Time) {
	if w.opt.Rotate {
		fpath := filepath.Join(w.dir, w.filename)
		fname := ""
		format := t.Format(w.RotateFormat)
		if format != w.lastRotateFormat {
			if w.opt.RotateMinutely {
				fname = fmt.Sprintf("%s.%s", w.filename, t.Add(-time.Minute).Format(w.RotateFormat))
			} else if w.opt.RotateHourly {
				fname = fmt.Sprintf("%s.%s", w.filename, t.Add(-time.Hour).Format(w.RotateFormat))
			} else if w.opt.RotateDaily {
				fname = fmt.Sprintf("%s.%s", w.filename, t.AddDate(0, 0, -1).Format(w.RotateFormat))
			}
		}

		if fname != "" {
			w.fileHandler.Close()
			if err := os.Rename(fpath, filepath.Join(w.dir, fname)); err != nil {
				w.stdLog.Printf("Rename error: %s", err)
				return
			}

			var err error
			w.fileHandler, err = NewFdHandler(fpath)
			if err != nil {
				w.stdLog.Printf("NewFdHandler error: %s", err)
			}
			w.lastRotateFormat = format
		}
	}
	return
}

func (w *LogFileWriter) putBuf(buf *bytes.Buffer) {
	buf.Reset()
	w.pool.Put(buf)
}

func (w *LogFileWriter) getBuf() *bytes.Buffer {
	return w.pool.Get().(*bytes.Buffer)
}

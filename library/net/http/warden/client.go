package warden

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zombie-k/kylin/library/libutil/hash"
	"github.com/zombie-k/kylin/library/net/netutil/breaker"
	xtime "github.com/zombie-k/kylin/library/time"
	"io"
	"net"
	xhttp "net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	_minRead = 16 * 1024
)

var (
	_noKickUserAgent = "kylin"
)

func init() {
	n, err := os.Hostname()
	if err == nil {
		_noKickUserAgent = _noKickUserAgent + runtime.Version() + " " + n
	}
}

type TAuth2 struct {
	Token  string
	Secret string
}

// ClientConfig is http client config.
type ClientConfig struct {
	Dial      xtime.Duration
	Timeout   xtime.Duration
	KeepAlive xtime.Duration
	Breaker   *breaker.Config
	URL       map[string]*ClientConfig
	Host      map[string]*ClientConfig
}

type Client struct {
	TAuth2
	conf      *ClientConfig
	client    *xhttp.Client
	dialer    *net.Dialer
	transport xhttp.Transport

	urlConf  map[string]*ClientConfig
	hostConf map[string]*ClientConfig
	mutex    sync.RWMutex
	breaker  *breaker.Group
}

func NewClient(c *ClientConfig) *Client {
	client := &Client{
		conf:     c,
		urlConf:  make(map[string]*ClientConfig, 0),
		hostConf: make(map[string]*ClientConfig, 0),
		breaker:  breaker.NewGroup(c.Breaker),
	}

	client.dialer = &net.Dialer{
		Timeout:   time.Duration(c.Dial),
		KeepAlive: time.Duration(c.KeepAlive),
	}

	client.client = &xhttp.Client{}
	if c.Timeout <= 0 {
		panic("must config http timeout!!!")
	}
	for uri, cfg := range c.URL {
		client.urlConf[uri] = cfg
	}
	for host, cfg := range c.Host {
		client.hostConf[host] = cfg
	}
	return client
}

func (client *Client) SetConfig(c *ClientConfig) {
	client.mutex.Lock()
	if c.Timeout > 0 {
		client.conf.Timeout = c.Timeout
	}
	if c.KeepAlive > 0 {
		client.conf.KeepAlive = c.KeepAlive
		client.dialer.KeepAlive = time.Duration(c.KeepAlive)
	}
	if c.Dial > 0 {
		client.conf.Dial = c.Dial
		client.dialer.Timeout = time.Duration(c.Timeout)
	}
	if c.Breaker != nil {
		client.conf.Breaker = c.Breaker
		client.breaker.Reload(c.Breaker)
	}
	for uri, cfg := range c.URL {
		client.urlConf[uri] = cfg
	}
	for host, cfg := range c.Host {
		client.hostConf[host] = cfg
	}
	client.mutex.Unlock()
}

func (client *Client) NewRequest(method, uri string, params url.Values) (req *xhttp.Request, err error) {
	if method == xhttp.MethodGet {
		req, err = xhttp.NewRequest(xhttp.MethodGet, fmt.Sprintf("%s?%s", uri, params.Encode()), nil)
	} else {
		req, err = xhttp.NewRequest(xhttp.MethodPost, uri, strings.NewReader(params.Encode()))
	}
	if err != nil {
		err = errors.New(fmt.Sprintf("%s,method:%s,uri:%s", err, method, uri))
		return
	}

	if method == xhttp.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("User-Agent", _noKickUserAgent)
	return
}

func (client *Client) Get(c context.Context, uri string, params url.Values, res interface{}) (err error) {
	req, err := client.NewRequest(xhttp.MethodGet, uri, params)
	if err != nil {
		return
	}
	return client.Do(c, req, res)
}

func (client *Client) Do(c context.Context, req *xhttp.Request, res interface{}, v ...string) (err error) {
	var bs []byte
	if bs, err = client.Raw(c, req, v...); err != nil {
		return err
	}
	if res != nil {
		if err = json.Unmarshal(bs, res); err != nil {
			err = fmt.Errorf("%s. host:%s", err, req.URL.Host)
		}
	}
	return
}

func (client *Client) Raw(c context.Context, req *xhttp.Request, v ...string) (bs []byte, err error) {
	var (
		code    string
		cancel  func()
		resp    *xhttp.Response
		config  *ClientConfig
		timeout time.Duration
		uri     = fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.Host, req.URL.Path)
	)
	// breaker
	brk := client.breaker.Get(uri)
	if err = brk.Allow(); err != nil {
		code = "breaker"
		_metricClientRequestCodeTotal.Inc(uri, req.Method, code)
		return
	}
	defer client.alterBreaker(brk, &err)
	// stat
	now := time.Now()
	defer func() {
		//_metricClientRequestDuration.Observe(int64(time.Since(now)/time.Millisecond), uri, req.Method)
		_metricClientRequestQuantile.Observe(int64(time.Since(now)/time.Millisecond), uri, req.Method)
		if code != "" {
			_metricClientRequestCodeTotal.Inc(uri, req.Method, code)
		}
	}()

	config = client.conf
	/*
		client.mutex.RLock()
		if config, ok = client.urlConf[uri]; !ok {
			if config, ok = client.hostConf[req.Host]; !ok {
				config = client.conf
			}
		}
		client.mutex.RUnlock()
	*/
	deliver := true
	timeout = time.Duration(config.Timeout)
	if deadline, ok := c.Deadline(); ok {
		if deadTimeout := time.Until(deadline); deadTimeout < timeout {
			timeout = deadTimeout
			deliver = false
		}
	}
	if deliver {
		c, cancel = context.WithTimeout(c, timeout)
		defer cancel()
	}
	setTimeout(req, timeout)
	req = req.WithContext(c)
	if resp, err = client.client.Do(req); err != nil {
		err = fmt.Errorf("%s host:%s, url:%s", err, req.URL.Host, realURL(req))
		code = "failed"
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= xhttp.StatusBadRequest {
		err = fmt.Errorf("incorret http status:%d host:%s, url:%s", resp.StatusCode, req.URL.Host, realURL(req))
		code = strconv.Itoa(resp.StatusCode)
		return
	}
	if bs, err = readBody(resp.Body, _minRead); err != nil {
		err = fmt.Errorf("%s host:%s, url:%s", err, req.URL.Host, realURL(req))
	}
	return
}

func (client *Client) alterBreaker(breaker breaker.Breaker, err *error) {
	if err != nil && *err != nil {
		breaker.MarkFailed()
	} else {
		breaker.MarkSuccess()
	}
}

func (client *Client) SignTAuth2(req *xhttp.Request, param, token, secret string) {
	token = url.QueryEscape(token)
	param = url.QueryEscape(param)
	digest, _ := hash.HmacSHA1(secret, param)
	sign := base64.RawStdEncoding.WithPadding(base64.StdPadding).EncodeToString(digest)
	sign = url.QueryEscape(sign)
	authorization := fmt.Sprintf("TAuth2 token=\"%s\", param=\"%s\", sign=\"%s\"", token, param, sign)
	req.Header.Add("Authorization", authorization)
}

func realURL(req *xhttp.Request) string {
	if req.Method == xhttp.MethodGet {
		return req.URL.String()
	} else if req.Method == xhttp.MethodPost {
		ru := req.URL.Path
		if req.Body != nil {
			rd, ok := req.Body.(io.Reader)
			if ok {
				buf := bytes.NewBuffer([]byte{})
				buf.ReadFrom(rd)
				ru = ru + "?" + buf.String()
			}
		}
		return ru
	}
	return req.URL.Path
}

func readBody(r io.Reader, capacity int64) (b []byte, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, capacity))
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}

package warden

import (
	"context"
	"github.com/pkg/errors"
	"github.com/zombie-k/kylin/library/net/http/warden/binding"
	"github.com/zombie-k/kylin/library/net/http/warden/render"
	"math"
	"net/http"
	"sync"
)

const (
	_abortIndex int8 = math.MaxInt8 / 2
)

var (
	_openParen  = []byte("(")
	_closeParen = []byte(")")
)

type Context struct {
	context.Context

	Request *http.Request
	Writer  http.ResponseWriter

	// flow control
	index    int8
	handlers []HandlerFunc

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]interface{}
	// This mutex protect Keys map
	keysMutex sync.RWMutex

	Error error

	method string
	engine *Engine

	RoutePath string
	Params    Params
}

/************************************/
/********** CONTEXT CREATION ********/
/************************************/
func (c *Context) reset() {
	c.Context = nil
	c.index = -1
	c.handlers = nil
	c.Keys = nil
	c.Error = nil
	c.method = ""
	c.RoutePath = ""
	c.Params = c.Params[0:0]
}

/************************************/
/*********** FLOW CONTROL ***********/
/************************************/
// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

// Abort prevents pending handlers from being called. Note that this will not stop current handler.
// For ex: you have an authorization middleware that validates that current request is authorized.
// If the authorization fails (ex: the password does not match), call Abort to ensure the remaining
// handlers for this request are not called.
func (c *Context) Abort() {
	c.index = _abortIndex
}

// AbortWithStatus calls `Abort()` and writes the headers with the specified status code.
// Ex, a failed attempt to authenticate a request could use: context.AbortWithStatus(401).
func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.Abort()
}

// IsAborted returns true if the current context was aborted.
func (c *Context) IsAborted() bool {
	return c.index >= _abortIndex
}

/************************************/
/******** METADATA MANAGEMENT********/
/************************************/

// Set is used to store a new key/value pair exclusively for this context.
// It also initializes c.Keys if it was not used previously.
func (c *Context) Set(key string, value interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
}

// Get returns the value for the given by key, ie: (value, true).
// If the value does not exists it returns (nil, false).
func (c *Context) Get(key string) (value interface{}, exists bool) {
	value, exists = c.Keys[key]
	return
}

/************************************/
/******** RESPONSE RENDERING ********/
/************************************/

// bodyAllowedForStatus is a copy of http.bodyAllowedForStatus non-exported function.
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 190:
		return false
	case status == 204:
		return false
	case status == 304:
		return false
	}
	return true
}

// Status sets the HTTP response code.
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

func (c *Context) Render(code int, r render.Render) {
	r.WriteContentType(c.Writer)
	if code > 0 {
		c.Status(code)
	}

	if !bodyAllowedForStatus(code) {
		return
	}

	params := c.Request.Form
	cb := params.Get("callback")
	jsonp := cb != "" && params.Get("jsonp") == "jsonp"
	if jsonp {
		c.Writer.Write([]byte(cb))
		c.Writer.Write(_openParen)
	}

	if err := r.Render(c.Writer); err != nil {
		c.Error = err
		return
	}

	if jsonp {
		if _, err := c.Writer.Write(_closeParen); err != nil {
			c.Error = errors.WithStack(err)
		}
	}
}

func (c *Context) JSON(code int, data interface{}, err error) {
	if code == 0 {
		code = http.StatusOK
	}
	c.Error = err
	c.Render(code, render.JSON{
		Code:    code,
		Message: "",
		Data:    data,
	})
}

func (c *Context) JSONMap(code int, data map[string]interface{}, err error) {
	if code == 0 {
		code = http.StatusOK
	}
	c.Error = err
	data["code"] = code
	/*
		if _, ok := data["message"]; !ok {
			data["message"] = ""
		}
	*/
	c.Render(code, render.MapJSON(data))
}

// String writes the given string into the response body.
func (c *Context) String(code int, format string, values ...interface{}) {
	c.Render(code, render.String{Format: format, Data: values})
}

func (c *Context) Bytes(code int, contentType string, data ...[]byte) {
	c.Render(code, render.Data{ContentType: contentType, Data: data})
}

func (c *Context) BindWith(obj interface{}, b binding.Binding) error {
	return c.mustBindWith(obj, b)
}

func (c *Context) Bind(obj interface{}) error {
	b := binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))
	return c.mustBindWith(obj, b)
}

func (c *Context) mustBindWith(obj interface{}, b binding.Binding) (err error) {
	if err = b.Bind(c.Request, obj); err != nil {
		c.Error = err
		c.Render(http.StatusOK, render.JSON{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data:    nil,
		})
		c.Abort()
	}
	return
}

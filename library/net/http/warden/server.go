package warden

import (
	"context"
	"github.com/pkg/errors"
	"github.com/zombie-k/kylin/library/net/metadata"
	xtime "github.com/zombie-k/kylin/library/time"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

// Handler responds to an HTTP request.
type Handler interface {
	ServeHTTP(c *Context)
}

// HandlerFunc http request handler function.
type HandlerFunc func(c *Context)

// ServeHTTP calls f(ctx).
func (f HandlerFunc) ServeHTTP(c *Context) {
	f(c)
}

type ServerConfig struct {
	Network      string
	Addr         string
	Timeout      xtime.Duration
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
}

type MethodConfig struct {
	Timeout xtime.Duration
}

type Engine struct {
	RouterGroup

	lock sync.RWMutex
	conf *ServerConfig

	address string

	trees methodTrees
	// *http.Server
	server    atomic.Value
	metastore map[string]map[string]interface{}

	pcLock        sync.RWMutex
	methodConfigs map[string]*MethodConfig

	//injections []injection

	// If enabled, the url.RawPath will be used to find parameters.
	UseRawPath bool

	// If true, the path value will be unescaped.
	// If UseRawPath is false (by default), the UnescapePathValues effectively is true,
	// as url.Path gonna be used, which is already unescaped.
	UnescapePathValues bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	allNoRoute  []HandlerFunc
	allNoMethod []HandlerFunc
	noMethod    []HandlerFunc
	noRoute     []HandlerFunc

	pool sync.Pool
}

// ServeHTTP confirms to the http.Handler interface.
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := engine.pool.Get().(*Context)
	c.Request = req
	c.Writer = w
	c.reset()

	engine.handleContext(c)
	engine.pool.Put(c)
}

func (engine *Engine) GetMethodConfig(path string) *MethodConfig {
	engine.pcLock.RLock()
	mc := engine.methodConfigs[path]
	engine.pcLock.RUnlock()
	return mc
}

func (engine *Engine) SetMethodConfig(path string, config *MethodConfig) {
	engine.pcLock.Lock()
	engine.methodConfigs[path] = config
	engine.pcLock.Unlock()
}

func (engine *Engine) SetConfig(conf *ServerConfig) (err error) {
	if conf.Timeout < 0 {
		return errors.New("warden: config timeout master greater than 0")
	}
	if conf.Network == "" {
		conf.Network = "tcp"
	}

	engine.lock.Lock()
	engine.conf = conf
	engine.lock.Unlock()
	return
}

func NewServer(conf *ServerConfig) *Engine {
	if conf == nil {
		//TODO:add default
		panic("config need")
	}

	engine := &Engine{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		address:                "127.0.0.1",
		trees:                  make(methodTrees, 0, 9),
		metastore:              make(map[string]map[string]interface{}),
		methodConfigs:          make(map[string]*MethodConfig),
		HandleMethodNotAllowed: true,
		pool:                   sync.Pool{},
	}

	if err := engine.SetConfig(conf); err != nil {
		panic(err)
	}
	engine.pool.New = func() interface{} {
		return engine.newContext()
	}
	engine.RouterGroup.engine = engine

	engine.addRoute(http.MethodGet, "/metrics", monitor())
	engine.NoRoute(func(c *Context) {
		c.Bytes(404, "text/plain", []byte("404 "+http.StatusText(404)))
		c.Abort()
	})
	engine.NoMethod(func(c *Context) {
		c.Bytes(405, "text/plain", []byte("405"+http.StatusText(405)))
		c.Abort()
	})

	return engine
}

func DefaultServer(conf *ServerConfig) *Engine {
	engine := NewServer(conf)
	engine.Use(Recovery(), Logger())
	return engine
}

func (engine *Engine) Start() error {
	conf := engine.conf
	l, err := net.Listen(conf.Network, conf.Addr)
	if err != nil {
		return errors.Wrapf(err, "warden: listen tcp: %s", conf.Addr)
	}

	server := &http.Server{
		ReadTimeout:  time.Duration(conf.ReadTimeout),
		WriteTimeout: time.Duration(conf.WriteTimeout),
	}

	go func() {
		//run server
		if err := engine.RunServer(server, l); err != nil {
			if errors.Cause(err) == http.ErrServerClosed {
				//log server closed
				return
			}
			panic(errors.Wrapf(err, "warden: engine.ListenServer(%+v, %+v)", server, l))
		}
	}()
	return nil
}

// RunServer will serve and start listening HTTP requests by give server and listener.
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine) RunServer(server *http.Server, l net.Listener) (err error) {
	server.Handler = engine
	engine.server.Store(server)
	if err = server.Serve(l); err != nil {
		err = errors.Wrapf(err, "listen server: %+v/%+v", server, l)
		return
	}
	return
}

func (engine *Engine) RunTLS(addr string, certFile string, keyFile string) (err error) {
	server := &http.Server{
		Addr:    addr,
		Handler: engine,
	}
	engine.server.Store(server)
	if err = server.ListenAndServeTLS(certFile, keyFile); err != nil {
		err = errors.Wrapf(err, "tls: %s/%s:%s", addr, certFile, keyFile)
	}
	return
}

func (engine *Engine) RunUnix(file string) (err error) {
	os.Remove(file)
	listener, err := net.Listen("unix", file)
	if err != nil {
		err = errors.Wrapf(err, "unix: %s", file)
		return
	}
	defer listener.Close()
	server := &http.Server{
		Handler: engine,
	}
	engine.server.Store(server)
	if err = server.Serve(listener); err != nil {
		err = errors.Wrapf(err, "unix: %s", err)
	}
	return
}

func (engine *Engine) addRoute(method string, path string, handlers ...HandlerFunc) {
	if path[0] != '/' {
		panic("warden: path must begin with '/'")
	}

	if method == "" {
		panic("warden: HTTP method can not be empty")
	}

	if len(handlers) == 0 {
		panic("warden: there must be at least one handler")
	}

	if _, ok := engine.metastore[path]; !ok {
		engine.metastore[path] = make(map[string]interface{})
	}
	engine.metastore[path]["method"] = method
	root := engine.trees.get(method)
	if root == nil {
		root = new(node)
		engine.trees = append(engine.trees, methodTree{method: method, root: root})
	}

	prelude := func(c *Context) {
		c.method = method
		c.RoutePath = path
	}
	handlers = append([]HandlerFunc{prelude}, handlers...)
	root.addRoute(path, handlers)
}

// Retrieve the routing handler and param via path
func (engine *Engine) prepareHandler(c *Context) {
	httpMethod := c.Request.Method
	rPath := c.Request.URL.Path
	unescape := false
	if engine.UseRawPath && len(c.Request.URL.EscapedPath()) > 0 {
		rPath = c.Request.URL.EscapedPath()
		unescape = engine.UnescapePathValues
	}
	rPath = pathClean(rPath)

	// Find root of the tree for the given HTTP method
	t := engine.trees
	for i, tl := 0, len(t); i < tl; i++ {
		if t[i].method != httpMethod {
			continue
		}
		root := t[i].root
		handlers, params, _ := root.getValue(rPath, c.Params, unescape)
		if handlers != nil {
			c.handlers = handlers
			c.Params = params
			return
		}
		break
	}

	if engine.HandleMethodNotAllowed {
		for _, tree := range engine.trees {
			if tree.method == httpMethod {
				continue
			}
			if handlers, _, _ := tree.root.getValue(rPath, nil, unescape); handlers != nil {
				c.handlers = engine.allNoMethod
				return
			}
		}
	}
	c.handlers = engine.allNoRoute
	return
}

func (engine *Engine) handleContext(c *Context) {
	var cancel func()
	req := c.Request
	contentType := req.Header.Get("Content-Type")
	switch {
	case strings.Contains(contentType, "multipart/form-data"):
		req.ParseMultipartForm(defaultMaxMemory)
	default:
		req.ParseForm()
	}

	// get derived timeout from http request header,
	// compare with the engine configured,
	// use the min one
	engine.lock.RLock()
	tm := time.Duration(engine.conf.Timeout)
	engine.lock.RUnlock()
	if dynC := engine.GetMethodConfig(req.URL.Path); dynC != nil {
		tm = time.Duration(dynC.Timeout)
	}
	if reqTm := timeout(req); reqTm > 0 && tm > reqTm {
		tm = reqTm
	}
	md := metadata.MD{
		metadata.RemoteIP:   remoteIp(req),
		metadata.RemotePort: remotePort(req),
	}
	parseMetadataTo(req, md)
	ctx := metadata.NewContext(context.Background(), md)
	if tm > 0 {
		c.Context, cancel = context.WithTimeout(ctx, tm)
	} else {
		c.Context, cancel = context.WithCancel(ctx)
	}
	defer cancel()
	engine.prepareHandler(c)
	c.Next()
}

func (engine *Engine) newContext() *Context {
	return &Context{engine: engine}
}

func (engine *Engine) Server() *http.Server {
	s, ok := engine.server.Load().(*http.Server)
	if !ok {
		return nil
	}
	return s
}

func (engine *Engine) Shutdown(ctx context.Context) error {
	server := engine.Server()
	if server == nil {
		return errors.New("warden: no server")
	}
	return errors.WithStack(server.Shutdown(ctx))
}

// Use attach a global middleware to the router. the middleware attached though Use() will be
// included in the handlers chain for every single request. Even 404, 405, static files...
// For example, this is the right place for a logger or error management middleware.
func (engine *Engine) Use(middleware ...Handler) IRoutes {
	engine.RouterGroup.Use(middleware...)
	engine.rebuild404Handlers()
	engine.rebuild405Handlers()
	return engine
}

func (engine *Engine) UseFunc(middleware ...HandlerFunc) IRoutes {
	engine.RouterGroup.UseFunc(middleware...)
	engine.rebuild404Handlers()
	engine.rebuild405Handlers()
	return engine
}

// Ping is used to set the general HTTP ping handler.
func (engine *Engine) Ping(handler HandlerFunc) {
	engine.GET("ping", handler)
}

// Register is used to export metadata to discovery.
func (engine *Engine) Register(handler HandlerFunc) {
	engine.GET("/register", handler)
}

func (engine *Engine) NoRoute(handlers ...HandlerFunc) {
	engine.noRoute = handlers
	engine.rebuild404Handlers()
}

func (engine *Engine) NoMethod(handlers ...HandlerFunc) {
	engine.noMethod = handlers
	engine.rebuild405Handlers()
}

func (engine *Engine) rebuild404Handlers() {
	engine.allNoRoute = engine.combineHandlers(engine.noRoute)
}

func (engine *Engine) rebuild405Handlers() {
	engine.allNoMethod = engine.combineHandlers(engine.noMethod)
}

func (engine *Engine) metadata() HandlerFunc {
	return func(c *Context) {
		//TODO:Render
	}
}

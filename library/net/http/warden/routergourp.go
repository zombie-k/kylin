package warden

import (
	"regexp"
)

// IRouter http router framework interface.
type IRouter interface {
	IRoutes
	Group(string, ...HandlerFunc) *RouterGroup
}

// IRoutes http router interface.
type IRoutes interface {
	UseFunc(...HandlerFunc) IRoutes
	Use(...Handler) IRoutes

	Handle(string, string, ...HandlerFunc) IRoutes
	HEAD(string, ...HandlerFunc) IRoutes
	GET(string, ...HandlerFunc) IRoutes
	POST(string, ...HandlerFunc) IRoutes
	PUT(string, ...HandlerFunc) IRoutes
	DELETE(string, ...HandlerFunc) IRoutes
}

type RouterGroup struct {
	Handlers   []HandlerFunc
	basePath   string
	engine     *Engine
	root       bool
	baseConfig *MethodConfig
}

var _ IRouter = &RouterGroup{}

func (group *RouterGroup) SetMethodConfig(config *MethodConfig) *RouterGroup {
	group.baseConfig = config
	return group
}

// Use adds middleware to the group.
func (group *RouterGroup) Use(middleware ...Handler) IRoutes {
	for _, m := range middleware {
		group.Handlers = append(group.Handlers, m.ServeHTTP)
	}
	return group.returnObj()
}

// UseFunc adds middleware to the group.
func (group *RouterGroup) UseFunc(middleware ...HandlerFunc) IRoutes {
	group.Handlers = append(group.Handlers, middleware...)
	return group.returnObj()
}

func (group *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		Handlers: group.combineHandlers(handlers),
		basePath: group.calculateAbsPath(relativePath),
		engine:   group.engine,
		root:     false,
	}
}

// Handle registers a new request handle and middleware with the given path and method.
// The last handler should be the real handler, the other ones should be middleware that can and should be shared among different routes.
// See the example code in doc.
//
// For HEAD, GET, POST, PUT, and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy)
func (group *RouterGroup) Handle(httpMethod string, relativePath string, handlers ...HandlerFunc) IRoutes {
	if matches, err := regexp.MatchString("^[A-Z]+$", httpMethod); !matches || err != nil {
		panic("http method" + httpMethod + " is not valid")
	}
	return group.handle(httpMethod, relativePath, handlers...)
}

// GET is a shortcut for router.Handle("GET", path, handle).
func (group *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("GET", relativePath, handlers...)
}

// POST is a shortcut for router.Handle("POST", path, handle).
func (group *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("POST", relativePath, handlers...)
}

// PUT is a shortcut for router.Handle("PUT", path, handle).
func (group *RouterGroup) PUT(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("PUT", relativePath, handlers...)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle).
func (group *RouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("DELETE", relativePath, handlers...)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle).
func (group *RouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("PATCH", relativePath, handlers...)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle).
func (group *RouterGroup) OPTIONS(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("OPTIONS", relativePath, handlers...)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle).
func (group *RouterGroup) HEAD(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("HEAD", relativePath, handlers...)
}

func (group *RouterGroup) handle(httpMethod string, relativePath string, handlers ...HandlerFunc) IRoutes {
	absolutePath := group.calculateAbsPath(relativePath)
	handlers = group.combineHandlers(handlers)
	group.engine.addRoute(httpMethod, absolutePath, handlers...)
	if group.baseConfig != nil {
		group.engine.SetMethodConfig(absolutePath, group.baseConfig)
	}
	return nil
}

func (group *RouterGroup) returnObj() IRoutes {
	if group.root {
		return group.engine
	}
	return group
}

func (group *RouterGroup) calculateAbsPath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)
}

func (group *RouterGroup) combineHandlers(handlerGroups ...[]HandlerFunc) []HandlerFunc {
	finalSize := len(group.Handlers)
	for _, handlers := range handlerGroups {
		finalSize += len(handlers)
	}
	if finalSize >= int(_abortIndex) {
		panic("too many handlers")
	}
	mergedHandlers := make([]HandlerFunc, finalSize)
	copy(mergedHandlers, group.Handlers)
	position := len(group.Handlers)
	for _, handlers := range handlerGroups {
		copy(mergedHandlers[position:], handlers)
		position += len(handlers)
	}
	return mergedHandlers
}

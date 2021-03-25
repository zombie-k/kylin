package warden

import (
	"net/url"
	"strings"
)

type Param struct {
	Key   string
	Value string
}

type Params []Param

type methodTree struct {
	method string
	root   *node
}

type methodTrees []methodTree

func (trees methodTrees) get(method string) *node {
	for _, tree := range trees {
		if tree.method == method {
			return tree.root
		}
	}
	return nil
}

const (
	static nodeType = iota
	root
	param
)

type nodeType uint32

type node struct {
	path      string
	indices   string
	children  []*node
	handlers  []HandlerFunc
	priority  uint32
	nType     nodeType
	maxParams uint8
	wildChild bool
}

func min(a int, b int) int {
	if a <= b {
		return a
	}
	return b
}

func countParams(path string) uint8 {
	n := 0

	for i := 0; i < len(path); i++ {
		if path[i] != ':' {
			continue
		}
		n++
	}

	if n >= 255 {
		return 255
	}

	return uint8(n)
}

func (n *node) incrementPriority(pos int) int {
	n.children[pos].priority++
	prio := n.children[pos].priority

	newPos := pos
	for newPos > 0 && n.children[newPos-1].priority < prio {
		n.children[newPos-1], n.children[newPos] = n.children[newPos], n.children[newPos-1]
		newPos--
	}

	n.indices = n.indices[:newPos] +
		n.indices[pos:pos+1] +
		n.indices[newPos:pos] +
		n.indices[pos+1:]

	return newPos
}

func (n *node) addRoute(path string, handlers []HandlerFunc) {
	fullPath := path
	n.priority++

	numParams := countParams(path)

	if len(n.path) > 0 || len(n.children) > 0 {
	walk:
		for {
			if numParams > n.maxParams {
				n.maxParams = numParams
			}

			//find common prefix
			i := 0
			max := min(len(n.path), len(path))
			for i < max && n.path[i] == path[i] {
				i++
			}

			if i < len(n.path) {
				child := &node{
					path:      n.path[i:],
					wildChild: n.wildChild,
					indices:   n.indices,
					children:  n.children,
					handlers:  n.handlers,
					priority:  n.priority - 1,
				}

				for _, ch := range n.children {
					if ch.maxParams > numParams {
						child.maxParams = ch.maxParams
					}
				}

				n.children = []*node{child}
				n.indices = string([]byte{n.path[i]})
				n.path = n.path[:i]
				n.handlers = nil
				n.wildChild = false
			}

			if i < len(path) {
				path = path[i:]

				if n.wildChild {
					n = n.children[0]
					n.priority++

					// Update maxParams of the child node
					if numParams > n.maxParams {
						n.maxParams = numParams
					}
					numParams--

					// Check if wildcard matches
					if len(path) >= len(n.path) && n.path == path[:len(n.path)] {
						if len(n.path) >= len(path) || path[len(n.path)] == '/' {
							continue walk
						}
					}

					pathSeg := strings.SplitN(path, "/", 2)[0]
					prefix := fullPath[:strings.Index(fullPath, pathSeg)] + n.path
					panic("'" + pathSeg +
						"' in new path '" + fullPath +
						"' conflicts with existing wildcard '" + n.path +
						"' in existing prefix '" + prefix +
						"'")
				}

				c := path[0]

				// Slash after param
				if n.nType == param && c == '/' && len(n.children) == 1 {
					n = n.children[0]
					n.priority++
					continue walk
				}

				for i := 0; i < len(n.indices); i++ {
					if c == n.indices[i] {
						i = n.incrementPriority(i)
						n = n.children[i]
						continue walk
					}
				}

				if c != ':' {
					child := &node{
						maxParams: numParams,
					}
					n.indices = n.indices + string([]byte{c})
					n.children = append(n.children, child)
					n.incrementPriority(len(n.indices) - 1)
					n = child
				}
				n.insertWildChild(numParams, path, fullPath, handlers)
				return
			} else if i == len(path) {
				if n.handlers != nil {
					panic("handlers are already registered for path '" + fullPath + "'")
				}
				n.handlers = handlers
			}
			return
		}
	} else {
		n.insertWildChild(numParams, path, fullPath, handlers)
		n.nType = root
	}
}

func (n *node) insertWildChild(numParams uint8, path string, fullPath string, handlers []HandlerFunc) {
	offset := 0

	//find prefix until first wildcard (beginning with ':')
	for i, max := 0, len(path); numParams > 0; i++ {
		c := path[i]
		if c != ':' {
			continue
		}

		//find wildcard end (either '/' or path end)
		end := i + 1
		for max < end && path[end] != '/' {
			switch path[end] {
			//the wildcard name must not contain ':'
			case ':':
				panic("only one wildcard per path segment is allowed, has: '" +
					path[i:] + "' in path '" + fullPath + "'")
			default:
				end++
			}
		}

		//check if this node existing children which would be unreachable
		//if we insert the wildcard here
		if len(n.children) > 0 {
			panic("wildcard route '" + path[i:end] +
				"' conflicts with existing children in path '" + fullPath + "'")
		}

		//check if the wildcard has a name
		if end-i < 2 {
			panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
		}

		if c == ':' {
			if i > 0 {
				n.path = path[offset:i]
				offset = i
			}

			child := &node{
				nType:     param,
				maxParams: numParams,
			}
			n.children = []*node{child}
			n.wildChild = true
			n = child
			n.priority++
			numParams--

			// if the path doesn't end with then wildcard, then
			// there will be another non-wildcard subpath starting with '/'
			if end < max {
				n.path = path[offset:end]
				offset = end

				child := &node{
					maxParams: numParams,
					priority:  1,
				}
				n.children = []*node{child}
				n = child
			}
		}
	}
	n.path = path[offset:]
	n.handlers = handlers
}

func (n *node) getValue(path string, para Params, unescape bool) (handlers []HandlerFunc, p Params, tsr bool) {
	p = para
walk:
	for {
		if len(path) > len(n.path) {
			if path[:len(n.path)] == n.path {
				path = path[len(n.path):]

				//无通配符,直接遍历child
				if !n.wildChild {
					c := path[0]
					for i := 0; i < len(n.indices); i++ {
						if c == n.indices[i] {
							n = n.children[i]
							continue walk
						}
					}

					//TODO:匹配失败
					return
				}

				//处理wildcard child
				n = n.children[0]
				switch n.nType {
				case param:
					// find param end ('/' or path end)
					end := 0
					for end < len(path) && path[end] != '/' {
						end++
					}

					// save param value
					if cap(p) < int(n.maxParams) {
						p = make(Params, n.maxParams)
					}
					i := len(p)
					p = p[:i+1]
					p[i].Key = n.path[1:]
					value := path[:end]
					if unescape {
						var err error
						if p[i].Value, err = url.QueryUnescape(value); err != nil {
							p[i].Value = value
						}
					} else {
						p[i].Value = value
					}

					// if has subpath
					if end < len(path) {
						if len(n.children) > 0 {
							path = path[end:]
							n = n.children[0]
							continue walk
						}

						//TODO: end with '/'
						return
					}

					if handlers = n.handlers; handlers != nil {
						return
					}

					if len(n.children) == 1 {
						n = n.children[0]
						tsr = n.path == "/" && n.handlers != nil
					}

					return
				default:
					panic("invalid node type")
				}
			}
		} else if path == n.path {
			if handlers = n.handlers; handlers != nil {
				return
			}

			if path == "/" && n.wildChild && n.nType != root {
				tsr = true
				return
			}

			for i := 0; i < len(n.indices); i++ {
				if n.indices[i] == '/' {
					n = n.children[i]
					tsr = len(n.path) == 1 && n.handlers != nil
					return
				}
			}

			return
		}

		tsr = (path == "/") ||
			(len(n.path) == len(path)+1 && n.path[len(path)] == '/' &&
				path == n.path[:len(n.path)-1] && n.handlers != nil)
		return
	}
}

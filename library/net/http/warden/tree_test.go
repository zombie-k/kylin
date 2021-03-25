package warden

import (
	"fmt"
	"testing"
)

func scanNode(root *node, level int, res *string) {
	tab := ""
	for i := 0; i < level; i++ {
		tab += "\t"
	}
	if level == 1 {
		fmt.Printf("%v\n", root)
		*res += fmt.Sprintf("%v\n", root)
	}
	if len(root.children) == 0 {
		return
	}
	for _, n := range root.children {
		fmt.Printf("%s%v\n", tab, n)
		*res += fmt.Sprintf("%s%v\n", tab, n)
		scanNode(n, level+1, res)
	}
	return
}

func TestTreeAdd(t *testing.T) {
	handlers := []HandlerFunc{func(c *Context) {
		fmt.Println("handler")
	}}
	root := new(node)
	root.addRoute("/metrics", handlers)
	root.addRoute("/metadata", handlers)
	root.addRoute("/metaable", handlers)
	//root.addRoute("/test", handlers)
	res := ""
	scanNode(root, 1, &res)
}

package main

import (
	"fmt"
	"github.com/weibocom/motan-go"
)

func main() {
	runClientDemo()
}

func runClientDemo() {
	mccontext := motan.GetClientContext("/Users/xiangqian5/github/kylin/app/example/motan/client/clientdemo.yaml")
	mccontext.Start(nil)
	//mclient := mccontext.GetClient("mytest-motan2")
	mclient := mccontext.GetClient("alchemy-rpc-refer")

	args := make(map[string]string, 16)
	args["name"] = "ray"
	args["id"] = "xxxx"
	var reply string
	err := mclient.Call("getConfigJson", []interface{}{args}, &reply)
	if err != nil {
		fmt.Printf("motan call fail! err:%v\n", err)
	} else {
		fmt.Printf("motan call success! reply:%s\n", reply)
	}
}
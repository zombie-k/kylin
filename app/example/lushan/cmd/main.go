package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Knetic/govaluate"
	jsoniter "github.com/json-iterator/go"
	"github.com/zombie-k/kylin/app/example/lushan/conf"
	"github.com/zombie-k/kylin/library/cache/memcache"
	"github.com/zombie-k/kylin/library/log"
	"strconv"
)

var (
	_key  string
	_db   int
	_iter bool
	_zlib bool
	_json bool
)

func init() {
	flag.StringVar(&_key, "key", "", "query key")
	flag.IntVar(&_db, "db", -1, "lushan db")
	flag.BoolVar(&_iter, "iter", false, "iter result list")
	flag.BoolVar(&_zlib, "zlib", false, "zlib compress/decompress")
	flag.BoolVar(&_json, "json", false, "json format")
}

func main() {
	flag.Parse()
	log.Init(conf.Conf.Log)
	defer log.Close()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	lushan()
}

func lushan() {
	memSlotsConf := conf.Conf.MemcacheSlots
	memConf := memSlotsConf.Memcache
	addrs := memSlotsConf.Addrs
	if len(addrs) == 0 {
		panic("Missing addrs in configuration.")
	}
	if _key == "" {
		panic("Missing query key, use '-k key'")
	}
	db := memSlotsConf.DB
	if _db != -1 {
		db = _db
	}
	if db <= 0 {
		panic("Missing db. use '-db db'")
	}

	config := memConf
	slot := hash()
	addr := addrs[slot : slot+1]
	config.Addr = fmt.Sprintf("%s:9764", addr)
	pool := memcache.New(config)
	query := fmt.Sprintf("%d-%s", db, _key)
	val := pool.Get(context.TODO(), query)
	if val == nil {
		return
	}
	if val.Item() == nil {
		fmt.Println(val.Item())
		return
	}
	if memSlotsConf.Zlib || _zlib {
		val.Item().Flags |= memcache.FlagZlib
	}
	parse(val)
}

func parse(val *memcache.Reply) {
	if val.Item().Flags == 0 {
		fmt.Println(string(val.Item().Value))
		return
	}

	var v []byte
	err := val.Scan(&v)
	if err != nil {
		fmt.Println("scan", err)
	}
	var vMap interface{}
	if _json {
		err = json.Unmarshal(v, &vMap)
		if err != nil {
			fmt.Println("unmarshal", err)
			return
		}
	}

	if vMap != nil && _iter {
		count := 0
		for k, item := range vMap.([]interface{}) {
			b, _ := jsoniter.Marshal(item)
			fmt.Println(string(b))
			count = k
		}
		fmt.Println("count:", count)
	} else {
		if vMap != nil {
			b, _ := jsoniter.Marshal(vMap)
			fmt.Println(string(b))
		} else {
			fmt.Println(string(v))
		}
	}
}

func hash() int {
	evalFunc := map[string]govaluate.ExpressionFunction{
		"substr": func(args ...interface{}) (interface{}, error) {
			str := args[0].(string)
			start := 0
			stop := len(str)
			if len(args) == 2 {
				start = int(args[1].(float64))
			} else if len(args) >= 3 {
				start = int(args[1].(float64))
				stop = int(args[2].(float64))
			}
			if start < 0 {
				start += len(str)
			}
			if stop < 0 {
				stop += len(str)
			}
			if start > stop {
				panic("startIndex must <= stopIndex")
			}
			return str[start:stop], nil
		},
		"mod": func(args ...interface{}) (interface{}, error) {
			if len(args) < 2 {
				panic("mod must have integer and num args")
			}
			str := args[0].(string)
			integer, err := strconv.Atoi(str)
			if err != nil {
				panic("mod error " + err.Error())
			}
			m := int(args[1].(float64))
			slot := integer % m
			return slot, nil
		},
	}
	hashExpression := conf.Conf.MemcacheSlots.HashExpr
	expr, _ := govaluate.NewEvaluableExpressionWithFunctions(hashExpression, evalFunc)
	param := make(map[string]interface{})
	param["key"] = _key
	result, _ := expr.Evaluate(param)
	return result.(int)
}

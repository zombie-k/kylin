package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/Knetic/govaluate"
	jsoniter "github.com/json-iterator/go"
	"github.com/zombie-k/kylin/app/example/redis/conf"
	"github.com/zombie-k/kylin/library/cache/redis"
	"github.com/zombie-k/kylin/library/container/pool"
	"github.com/zombie-k/kylin/library/libutil/zlib"
	xtime "github.com/zombie-k/kylin/library/time"
	"log"
	"strconv"
	"time"
)

var (
	_RedisMapKey = "%s_%d"
	_QRedisName  string
	_QKey        string
	_QRedisDB    int
	redisMap     = make(map[string]*RedisMap)
	compress     = zlib.New()
	_zlib        bool
	_json        bool
	_suffix      string
)

type RedisMap struct {
	poolSlice []*redis.Pool
	hashExpr  string
	zlib      bool
	json      bool
}

func init() {
	flag.StringVar(&_QRedisName, "n", "", "name of redis set in config file")
	flag.StringVar(&_QKey, "k", "", "redis key")
	flag.IntVar(&_QRedisDB, "d", 0, "redis db. default 0")
}

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	if _QRedisName == "" {
		fmt.Println("need redis name.")
		return
	}
	if _QKey == "" {
		fmt.Println("need query key.")
		return
	}
	conn, err := connTarget()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	exec2(conn)
}

func exec2(conn redis.Conn) {
	val, err := redis.Bytes(conn.Do("GET", _QKey+_suffix))
	if _zlib {
		val, err = compress.UnCompress(val)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	if _json {
		var v interface{}
		_ = jsoniter.Unmarshal(val, &v)
		val, _ = jsoniter.Marshal(v)
	}
	fmt.Println(string(val))
}

func exec() {
	mapKey := fmt.Sprintf(_RedisMapKey, _QRedisName, _QRedisDB)
	if info, ok := redisMap[mapKey]; ok {
		var p *redis.Pool
		if len(info.poolSlice) > 1 {
			idx := hash(info.hashExpr, _QKey)
			p = info.poolSlice[idx]
		} else {
			p = info.poolSlice[0]
		}
		if p == nil {
			panic("pool nil")
		}
		conn := p.Get(context.TODO())
		defer conn.Close()
		val, err := redis.Bytes(conn.Do("GET", _QKey))
		if info.zlib {
			val, err = compress.UnCompress(val)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		if info.json {
			var v interface{}
			_ = jsoniter.Unmarshal(val, &v)
			val, _ = jsoniter.Marshal(v)
		}
		fmt.Println(string(val))
	}
}

func connTarget() (redis.Conn, error) {
	redisConf := conf.Conf.Redis
	for i, _ := range redisConf {
		config := redisConf[i]
		if config.Name == _QRedisName {
			var addr string
			if len(config.Addrs) > 1 {
				idx := hash(config.HashExpr, _QKey)
				addr = config.Addrs[idx]
			} else if len(config.Addrs) == 1 {
				addr = config.Addrs[0]
			}
			if addr != "" {
				rdxConfig := &redis.Config{
					Config: &pool.Config{
						Active: 2,
						Idle:   2,
					},
					Name:         config.Name,
					Proto:        "tcp",
					Addr:         addr,
					Db:           config.DB,
					DialTimeout:  xtime.Duration(5 * time.Second),
					ReadTimeout:  xtime.Duration(5 * time.Second),
					WriteTimeout: xtime.Duration(5 * time.Second),
				}
				_zlib = config.Zlib
				_json = config.Json
				_suffix = config.Suffix
				return redis.NewConn(rdxConfig)
			}
		}
	}
	return nil, errors.New("error")
}

func connAll() {
	redisConf := conf.Conf.Redis
	for i, _ := range redisConf {
		config := redisConf[i]
		if config.Name == "" {
			log.Printf("Missing name config for addrs:%+v\n", config.Addrs)
			continue
		}
		if len(config.Addrs) == 0 {
			log.Printf("Missing addrs config for name:%s\n", config.Name)
			continue
		}
		active := config.Active
		if config.Active == 0 {
			active = 5
		}
		idle := config.Idle
		if config.Idle == 0 {
			idle = 5
		}
		if idle > active {
			idle = active
		}
		for _, addr := range config.Addrs {
			rdxConfig := &redis.Config{
				Config: &pool.Config{
					Active: active,
					Idle:   idle,
				},
				Name:         config.Name,
				Proto:        "tcp",
				Addr:         addr,
				Db:           config.DB,
				DialTimeout:  xtime.Duration(5 * time.Second),
				ReadTimeout:  xtime.Duration(5 * time.Second),
				WriteTimeout: xtime.Duration(5 * time.Second),
			}
			mapKey := fmt.Sprintf(_RedisMapKey, config.Name, config.DB)
			if rm, ok := redisMap[mapKey]; ok {
				rm.poolSlice = append(rm.poolSlice, redis.NewPool(rdxConfig))
			} else {
				redisMap[mapKey] = &RedisMap{
					poolSlice: []*redis.Pool{redis.NewPool(rdxConfig)},
					hashExpr:  config.HashExpr,
					zlib:      config.Zlib,
					json:      config.Json,
				}
			}
		}
	}
}

func hash(hashExpr string, key string) int {
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
	expr, _ := govaluate.NewEvaluableExpressionWithFunctions(hashExpr, evalFunc)
	param := make(map[string]interface{})
	param["key"] = key
	result, _ := expr.Evaluate(param)
	return result.(int)
}

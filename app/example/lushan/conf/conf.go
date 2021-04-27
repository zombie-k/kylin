package conf

import (
	"errors"
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/zombie-k/kylin/library/cache/memcache"
	"github.com/zombie-k/kylin/library/log"
)

var (
	confPath string
	Conf     = &Config{}
)

type Config struct {
	Log           *log.Config
	MemcacheSlots *memcacheSlots
}

type memcacheSlots struct {
	Memcache *memcache.Config
	Addrs    []string
	DB       int
	HashExpr string
	Zlib     bool
	Json     bool
}

func init() {
	flag.StringVar(&confPath, "c", "", "default config path")
}

func Init() error {
	if confPath == "" {
		return errors.New("confPath nil")
	}
	_, err := toml.DecodeFile(confPath, &Conf)
	return err
}

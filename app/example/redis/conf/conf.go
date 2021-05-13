package conf

import (
	"flag"
	"github.com/BurntSushi/toml"
)

type Config struct {
	Redis []*RdConfig
}

type RdConfig struct {
	Name string
	DB int
	Idle int
	Active int
	Addrs []string
	HashExpr string
	Zlib bool
	Json bool
	Suffix string
}

var (
	confPath string
	Conf = &Config{}
	_idle = 10
	_active = 10
	_db = 0
)

func init() {
	flag.StringVar(&confPath, "c", "", "default config path")
}

func Init() error {
	if confPath == "" {
		panic("Missing config path. Use '-h' for help")
	}
	_, err := toml.DecodeFile(confPath, &Conf)
	return err
}

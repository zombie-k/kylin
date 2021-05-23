package conf

import (
	"errors"
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/zombie-k/kylin/library/log"
	"github.com/zombie-k/kylin/library/pkg/kafka"
	xtime "github.com/zombie-k/kylin/library/time"
)

var (
	confPath string
	Conf     = &Config{}
)

func init() {
	flag.StringVar(&confPath, "c", "", "config path")
}

type Config struct {
	Kafka *kafka.Config
	Log *log.Config
	Core struct{
		Sleep xtime.Duration
	}
}

func Init() error {
	if confPath == "" {
		return errors.New("confPath nil")
	}
	if _, err := toml.DecodeFile(confPath, &Conf); err != nil {
		return err
	}
	err := Conf.Kafka.Builder()
	return err
}

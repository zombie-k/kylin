package main

import (
	"flag"
	"github.com/zombie-k/kylin/app/example/kafka/conf"
	"github.com/zombie-k/kylin/app/example/kafka/service"
	"github.com/zombie-k/kylin/library/log"
	"github.com/zombie-k/kylin/library/pkg/kafka"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}

	srv := service.New(conf.Conf)
	consumer, err := kafka.NewConsumer(conf.Conf, srv.Parser(conf.Conf.Kafka.Name))
	if err != nil {
		panic(err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			consumer.Stop()
			log.Info("exiting...")
			return
		case syscall.SIGHUP:
		default:
			consumer.Stop()
			return
		}
	}
}

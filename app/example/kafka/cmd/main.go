package main

import (
	"github.com/zombie-k/kylin/app/example/kafka/conf"
	"github.com/zombie-k/kylin/library/pkg/kafka"
	"os"
	"os/signal"
	"syscall"
)

func run() {
	if err := conf.Init(); err != nil {
		panic(err)
	}
	consumer, err := kafka.NewConsumer(conf.Conf)
	if err != nil {
		panic(err)
	}
	defer func() {
		consumer.Stop()
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP|syscall.SIGTERM|syscall.SIGINT|syscall.SIGQUIT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func main() {
	run()
}

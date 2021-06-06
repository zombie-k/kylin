package service

import (
	"fmt"
	"github.com/zombie-k/kylin/app/example/kafka/conf"
	"github.com/zombie-k/kylin/library/pkg/kafka"
)

type Service struct {
	c         *conf.Config
	parserMap map[string]func() kafka.Messager
	loaderMap map[string]func() kafka.Loader
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:         c,
		parserMap: make(map[string]func() kafka.Messager),
		loaderMap: make(map[string]func() kafka.Loader),
	}

	s.ServiceRegisterParser()
	s.ServiceRegisterLoader()

	return
}

func (s *Service) RegisterParser(name string, parser kafka.Messager) {
	if _, ok := s.parserMap[name]; ok {
		panic(fmt.Sprintf("%s already register", name))
	}
	s.parserMap[name] = func() kafka.Messager {
		return parser
	}
}

func (s *Service) Parser(name string) kafka.Messager {
	if f, ok := s.parserMap[name]; ok {
		return f()
	} else {
		panic(fmt.Sprintf("%s not register", name))
	}
}

func (s *Service) RegisterLoader(name string, loader kafka.Loader) {
	if _, ok := s.loaderMap[name]; ok {
		panic(fmt.Sprintf("%s already register", name))
	}
	s.loaderMap[name] = func() kafka.Loader {
		return loader
	}
}

func (s *Service) Loader(name string) kafka.Loader {
	if f, ok := s.loaderMap[name]; ok {
		return f()
	} else {
		panic(fmt.Sprintf("%s not register", name))
	}
}

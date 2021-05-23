package service

import (
	"fmt"
	"github.com/zombie-k/kylin/app/example/kafka/conf"
	"github.com/zombie-k/kylin/library/pkg/kafka"
)

type Service struct {
	c         *conf.Config
	parserMap map[string]func() kafka.Messager
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:         c,
		parserMap: make(map[string]func() kafka.Messager),
	}

	s.ServiceRegister()

	return
}

func (s *Service) Register(name string, parser kafka.Messager) {
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

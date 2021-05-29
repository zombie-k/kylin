package service

import (
	basic "github.com/zombie-k/kylin/app/example/kafka/service/parser"
	"github.com/zombie-k/kylin/app/example/kafka/service/parser/minimalism"
	"time"
)

func (s *Service) ServiceRegister() {
	s.Register(minimalism.Name, minimalism.Parser(basic.BasicSleepingTime(time.Duration(s.c.Core.Sleep))))
}

func (s *Service) ServiceRegisterLoader() {
	s.RegisterLoader(minimalism.Name, minimalism.Loader(s.c))
}

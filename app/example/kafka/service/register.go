package service

import (
	basic "github.com/zombie-k/kylin/app/example/kafka/service/parser"
	"github.com/zombie-k/kylin/app/example/kafka/service/parser/minimalism"
	"github.com/zombie-k/kylin/app/example/kafka/service/parser/video799tab"
	"time"
)

func (s *Service) ServiceRegisterParser() {
	s.RegisterParser(minimalism.Name, minimalism.Parser(basic.BasicSleepingTime(time.Duration(s.c.Core.Sleep))))
	s.RegisterParser(video799tab.Name, video799tab.Parser(s.c))
}

func (s *Service) ServiceRegisterLoader() {
	s.RegisterLoader(minimalism.Name, minimalism.Loader(s.c))
}

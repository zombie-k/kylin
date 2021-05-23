package service

import (
	basic "github.com/zombie-k/kylin/app/example/kafka/service/parser"
	"github.com/zombie-k/kylin/app/example/kafka/service/parser/video799tab"
	"time"
)

func (s *Service) ServiceRegister() {
	s.Register(video799tab.Name, video799tab.New(basic.BasicSleepingTime(time.Duration(s.c.Core.Sleep))))
}

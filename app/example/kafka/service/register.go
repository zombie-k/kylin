package service

import "github.com/zombie-k/kylin/app/example/kafka/service/video799tab"

func (s *Service) ServiceRegister() {
	s.Register(video799tab.Name, video799tab.New())
}

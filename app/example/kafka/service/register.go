package service

import (
	"github.com/zombie-k/kylin/app/example/kafka/service/parser/minimalism"
	"github.com/zombie-k/kylin/app/example/kafka/service/parser/video799tab"
	"github.com/zombie-k/kylin/app/example/kafka/service/parser/videotabissue"
)

func (s *Service) ServiceRegisterParser() {
	s.RegisterParser(minimalism.Name, minimalism.Parser(s.c))
	s.RegisterParser(video799tab.Name, video799tab.Parser(s.c))
	s.RegisterParser(videotabissue.Name, videotabissue.Parser(s.c))
}

func (s *Service) ServiceRegisterLoader() {
	s.RegisterLoader(minimalism.Name, minimalism.Loader(s.c))
}

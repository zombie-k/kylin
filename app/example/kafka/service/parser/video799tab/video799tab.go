package video799tab

import (
	basic "github.com/zombie-k/kylin/app/example/kafka/service/parser"
	"github.com/zombie-k/kylin/library/log"
	"github.com/zombie-k/kylin/library/pkg/kafka"
	xtime "github.com/zombie-k/kylin/library/time"
	"time"
)

const Name = "video799tab"

type parser struct{
	basic.Basic
}

func New(opts ...basic.BasicOption) *parser {
	basicOption := basic.Basic{}
	for _, opt := range opts {
		opt.F(&basicOption)
	}
	return &parser{basicOption}
}

func (p *parser) Messages(message *kafka.Message) {
	log.Info("%s %s %d %d %s %s", message.Timestamp.Format(xtime.LongDateFormatter), message.Topic, message.Partition, message.Offset, message.Key, message.Value)
	if p.Sleep > 0 {
		time.Sleep(p.Sleep)
	}
}

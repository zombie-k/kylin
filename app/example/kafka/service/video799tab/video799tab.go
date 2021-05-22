package video799tab

import (
	"github.com/zombie-k/kylin/library/log"
	"github.com/zombie-k/kylin/library/pkg/kafka"
	xtime "github.com/zombie-k/kylin/library/time"
)

const Name = "video799tab"

type parser struct{}

func New() *parser {
	return &parser{}
}

func (p *parser) Messages(message *kafka.Message) {
	log.Info("%s %s %d %d %s %s", message.Timestamp.Format(xtime.LongDateFormatter), message.Topic, message.Partition, message.Offset, message.Key, message.Value)
}

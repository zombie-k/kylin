package videotabissue

import (
	"github.com/zombie-k/kylin/library/log"
	"github.com/zombie-k/kylin/library/pkg/kafka"
	xtime "github.com/zombie-k/kylin/library/time"
)

const Name = "video_tab_issue"

type parser struct {
	configure kafka.Configure
}

func Parser(configure kafka.Configure) *parser {
	return &parser{configure: configure}
}

func (p *parser) Messages(message *kafka.Message) {
	log.Info("%s\t%s\t%d\t%d\t%d\t%d\t%s",
		message.Timestamp.Format(xtime.LongDateFormatter),
		message.Topic,
		message.Partition,
		message.Offset,
		message.HighWaterMarkOffset,
		message.HighWaterMarkOffset-message.Offset,
		message.Value)
	/*
		asleep := time.Duration(p.configure.GetConfig().(*conf.Config).Core.Sleep)
		if asleep > 0 {
			time.Sleep(asleep)
		}
	*/
}

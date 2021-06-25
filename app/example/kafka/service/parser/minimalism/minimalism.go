package minimalism

import (
	"context"
	"fmt"
	"github.com/zombie-k/kylin/app/example/kafka/conf"
	basic "github.com/zombie-k/kylin/app/example/kafka/service/parser"
	"github.com/zombie-k/kylin/library/log"
	"github.com/zombie-k/kylin/library/pkg/kafka"
	xtime "github.com/zombie-k/kylin/library/time"
	"time"
)

const Name = "minimalism"

type parser struct {
	configure kafka.Configure
}

func Parser(configure kafka.Configure) *parser {
	return &parser{configure: configure}
}

func (p *parser) Messages(message *kafka.Message) {
	//log.Info("%s %s %d %d %s %s", message.Timestamp.Format(xtime.LongDateFormatter), message.Topic, message.Partition, message.Offset, message.Key, message.Value)
	log.Info("%s %s %d %d %d %d", message.Timestamp.Format(xtime.LongDateFormatter), message.Topic, message.Partition, message.Offset, message.HighWaterMarkOffset, message.HighWaterMarkOffset-message.Offset)
	asleep := time.Duration(p.configure.GetConfig().(*conf.Config).Core.Sleep)
	if asleep > 0 {
		//time.Sleep(asleep)
	}
}

type loader struct {
	basic.Basic
	ch chan *kafka.Message
	c  *conf.Config
}

func Loader(c *conf.Config, opts ...basic.BasicOption) *loader {
	basicOption := basic.Basic{}
	for _, opt := range opts {
		opt.F(&basicOption)
	}
	ch := make(chan *kafka.Message, 100)
	return &loader{
		Basic: basicOption,
		ch:    ch,
		c:     c,
	}
}

func (l *loader) Inputs() <-chan *kafka.Message {
	return l.ch
}

func (l *loader) Start(ctx context.Context) {
	for i := 0; i < 1000; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			l.ch <- &kafka.Message{
				Value:     []byte(fmt.Sprintf("hello %d", i)),
				Topic:     l.c.Kafka.Produce.Topic,
				Partition: 2,
			}
		}
	}
}

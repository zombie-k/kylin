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
	basic.Basic
}

func Parser(opts ...basic.BasicOption) *parser {
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

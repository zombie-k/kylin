package kafka

import (
	"context"
	"github.com/zombie-k/kylin/library/log"
	xtime "github.com/zombie-k/kylin/library/time"
	"time"
)

type dealOption struct {
	asleep time.Duration
	parser func(context.Context, *Message)
}

type DealOptions struct {
	f func(*dealOption)
}

func DealFunction(fn func(ctx context.Context, msg *Message)) DealOptions {
	return DealOptions{f: func(option *dealOption) {
		option.parser = fn
	}}
}

func DealAsleep(d time.Duration) DealOptions {
	return DealOptions{func(option *dealOption) {
		option.asleep = d
	}}
}

type Processor struct {
	opt    *dealOption
	parser func(context.Context, *Message)
}

func DefaultProcessor(opts ...DealOptions) *Processor {
	option := dealOption{
		parser: defaultParser,
	}
	for _, opt := range opts {
		opt.f(&option)
	}
	processor := &Processor{
		opt:    &option,
		parser: option.parser,
	}
	return processor
}

func (p *Processor) Messages(message *Message) {
	p.parser(context.TODO(), message)
}

func defaultParser(ctx context.Context, message *Message) {
	log.Info("%s %s %d %d %s %s", message.Timestamp.Format(xtime.LongDateFormatter), message.Topic, message.Partition, message.Offset, message.Key, message.Value)
}

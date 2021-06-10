package kafka

import (
	"context"
	"github.com/zombie-k/kylin/library/log"
	xtime "github.com/zombie-k/kylin/library/time"
)

type dealOption struct {
	parser    func(context.Context, *Processor, *Message)
	configure Configure
}

type DealOptions struct {
	f func(*dealOption)
}

func DealFunction(fn func(ctx context.Context, p *Processor, msg *Message)) DealOptions {
	return DealOptions{f: func(option *dealOption) {
		option.parser = fn
	}}
}

func DealConfigure(configure Configure) DealOptions {
	return DealOptions{func(option *dealOption) {
		option.configure = configure
	}}
}

type Processor struct {
	opt    *dealOption
	parser func(context.Context, *Processor, *Message)

	Configurer Configure
}

func (p *Processor) Opt() *dealOption {
	return p.opt
}

func DefaultProcessor(opts ...DealOptions) *Processor {
	option := dealOption{
		parser: defaultParser,
	}
	for _, opt := range opts {
		opt.f(&option)
	}
	processor := &Processor{
		opt:        &option,
		parser:     option.parser,
		Configurer: option.configure,
	}
	return processor
}

func (p *Processor) Messages(message *Message) {
	p.parser(context.TODO(), p, message)
}

func defaultParser(ctx context.Context, p *Processor, message *Message) {
	log.Info("%s %s %d %d %s %s", message.Timestamp.Format(xtime.LongDateFormatter), message.Topic, message.Partition, message.Offset, message.Key, message.Value)
}

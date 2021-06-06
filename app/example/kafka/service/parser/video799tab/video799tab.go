package video799tab

import (
	"context"
	"github.com/zombie-k/kylin/library/log"
	"github.com/zombie-k/kylin/library/pkg/kafka"
	xtime "github.com/zombie-k/kylin/library/time"
)

const Name = "video799tab"

func Parser() *kafka.Processor {
	p := kafka.DefaultProcessor(kafka.DealFunction(parse))
	return p
}

func parse(ctx context.Context, message *kafka.Message) {
	log.Info("%s %s %d %d %s %s", message.Timestamp.Format(xtime.LongDateFormatter), message.Topic, message.Partition, message.Offset, message.Key, message.Value)
}

package video799tab

import (
	"context"
	"github.com/zombie-k/kylin/app/example/kafka/conf"
	"github.com/zombie-k/kylin/library/log"
	"github.com/zombie-k/kylin/library/pkg/kafka"
	xtime "github.com/zombie-k/kylin/library/time"
	"time"
)

const Name = "video799tab"

func Parser(c *conf.Config) *kafka.Processor {
	p := kafka.DefaultProcessor(kafka.DealFunction(parse), kafka.DealConfigure(c))
	return p
}

func parse(ctx context.Context, processor *kafka.Processor, message *kafka.Message) {
	log.Info("%s %s %d %d key:%s val:%s", message.Timestamp.Format(xtime.LongDateFormatter), message.Topic, message.Partition, message.Offset, message.Key, message.Value)
	/*
	v := &model.VideoTab799{}
	err := jsoniter.Unmarshal(message.Value, v)
	if err != nil {
		fmt.Println(err)
	}
	for _, item := range v.ExtraList {
		if _, ok := item["uid"].(string); ok {
			log.Info("%s %s %d %d key:%s val:%s", message.Timestamp.Format(xtime.LongDateFormatter), message.Topic, message.Partition, message.Offset, message.Key, message.Value)
		}
	}
	*/

	if c, ok := processor.Configurer.GetConfig().(*conf.Config); ok {
		time.Sleep(time.Duration(c.Core.Sleep))
	}
}

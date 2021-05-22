package video799tab

import (
	"fmt"
	"github.com/zombie-k/kylin/library/pkg/kafka"
	"time"
)

const Name = "video799tab"

type parser struct{}

func New() *parser {
	return &parser{}
}

func (p *parser) Messages(message *kafka.Message) {
	fmt.Printf("%s %s %d %d %s %s\n", message.Timestamp, message.Topic, message.Partition, message.Offset, message.Key, message.Value)
	time.Sleep(time.Second * 5)
}

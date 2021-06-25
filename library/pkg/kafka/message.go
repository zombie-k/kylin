package kafka

import (
	"context"
	"time"
)

type Messager interface {
	Messages(message *Message)
}

type Loader interface {
	Inputs() <-chan *Message
	Start(ctx context.Context)
}

type Message struct {
	// Message key
	Key []byte

	// Raw message
	Value []byte

	// Message offset from input
	Offset int64

	// Message partition from input
	Partition int32

	// Message topic
	Topic string

	// Message timestamp
	Timestamp time.Time

	// Message high water mark offset of the partition
	HighWaterMarkOffset int64
}

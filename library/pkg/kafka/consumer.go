package kafka

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/zombie-k/kylin/library/log"
	"github.com/zombie-k/kylin/library/sync/pipeline/fanout"
	"strings"
	"sync"
	"time"
)

type Consumer struct {
	c      *Config
	config *sarama.Config

	consumerGroup sarama.ConsumerGroup
	client        sarama.Client
	handle        *handle
	parser        Messager

	ctx    context.Context
	cancel func()

	job *fanout.Fanout
	wg  *sync.WaitGroup
}

func (consumer *Consumer) Stop() {
	consumer.cancel()
	consumer.wg.Wait()
}

func NewConsumer(c *Config, parser Messager) (consumer *Consumer, err error) {
	if parser == nil {
		parser = DefaultProcessor()
	}
	version, err := sarama.ParseKafkaVersion(c.Consume.Version)
	if err != nil {
		return nil, err
	}

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = version
	config.ClientID = fmt.Sprintf("%s-%s", c.Consume.Name, c.Consume.Group)

	switch strings.ToLower(c.Consume.OffsetMode) {
	case "oldest", "earliest":
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	case "latest", "newest":
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	default:
	}

	switch strings.ToLower(c.Consume.Rebalance) {
	case "range":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	case "sticky":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	case "roundrobin":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	default:
	}

	if c.Consume.Sasl.Enable {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = c.Consume.Sasl.User
		config.Net.SASL.Password = c.Consume.Sasl.Password
		switch strings.ToUpper(c.Consume.Sasl.Mechanism) {
		case sarama.SASLTypePlaintext:
			config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		case sarama.SASLExtKeyAuth:
			config.Net.SASL.Mechanism = sarama.SASLExtKeyAuth
		case sarama.SASLTypeGSSAPI:
			config.Net.SASL.Mechanism = sarama.SASLTypeGSSAPI
		case sarama.SASLTypeOAuth:
			config.Net.SASL.Mechanism = sarama.SASLTypeOAuth
		default:
			config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		}
	}

	/*
		client, err := sarama.NewConsumerGroup(c.Consume.Brokers, c.Consume.Group, config)
		if err != nil {
			return nil, err
		}
	*/
	client, err := sarama.NewClient(c.Consume.Brokers, config)
	if err != nil {
		return nil, err
	}
	consumeGroup, err := sarama.NewConsumerGroupFromClient(c.Consume.Group, client)
	if err != nil {
		_ = client.Close()
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	consumer = &Consumer{
		config:        config,
		client:        client,
		consumerGroup: consumeGroup,
		handle: &handle{
			ready:       make(chan bool, 0),
			name:        c.Consume.Name,
			logInterval: 10 * time.Second,
			wg:          &sync.WaitGroup{},
		},
		parser: parser,
		ctx:    ctx,
		cancel: cancel,
		wg:     &sync.WaitGroup{},
		job:    fanout.New("kafka", fanout.Worker(c.Consume.Job.Worker), fanout.Buffer(c.Consume.Job.Buffer)),
	}

	consumer.handle.consumer = consumer
	consumer.wg.Add(1)
	go func() {
		defer consumer.wg.Done()
		for {
			select {
			case <-consumer.ctx.Done():
				log.Warn("Terminating: context cancelled")
				return
			case err := <-consumer.consumerGroup.Errors():
				log.Error("%s", err.Error())
			default:
				if err := consumer.consumerGroup.Consume(consumer.ctx, c.Consume.Topics, consumer.handle); err != nil {
					switch err {
					case sarama.ErrClosedClient, sarama.ErrClosedConsumerGroup:
						log.Error("%v", err)
						return
					case sarama.ErrOutOfBrokers:
						log.Warn("%v", err)
					default:
						log.Warn("%v", err)
					}
				}
				if consumer.ctx.Err() != nil {
					log.Warn("Terminating:%v", ctx.Err())
					return
				}
				time.Sleep(time.Second)
				consumer.handle.ready = make(chan bool, 0)
			}
		}
	}()
	go func() {
		consumer.wg.Wait()
		if err := consumeGroup.Close(); err != nil {
			log.Error("Error closing consumerGroup: %v", err)
		}
	}()

	log.Debug("Topic:%s", c.Consume.Topics)
	log.Debug("Group:%s", c.Consume.Group)
	<-consumer.handle.ready
	log.Debug("Kylin consumer up and running!...\n")
	return
}

type handle struct {
	name        string
	logInterval time.Duration

	ready chan bool
	wg    *sync.WaitGroup

	consumer *Consumer
}

func (h *handle) Setup(session sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *handle) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *handle) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	logTick := time.NewTicker(h.logInterval)

	for {
		select {
		case <-session.Context().Done():
			log.Debug("topic:%s, partition:%d session exit, waiting and processing the buffers", claim.Topic(), claim.Partition())
			h.wg.Wait()
			return nil
		case <-logTick.C:
			log.Debug("topic:%s, partition:%d, initialOffset:%d, highWaterMarkOffset:%d, channel buffer size:%d",
				claim.Topic(),
				claim.Partition(),
				claim.InitialOffset(),
				claim.HighWaterMarkOffset(),
				h.consumer.job.Channel())
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			h.wg.Add(1)
			if err := h.consumer.job.DoWait(context.TODO(), func(ctx context.Context) {
				defer func() {
					h.wg.Done()
				}()
				message := &Message{
					Key:                 msg.Key,
					Value:               msg.Value,
					Offset:              msg.Offset,
					Partition:           msg.Partition,
					Topic:               msg.Topic,
					Timestamp:           msg.Timestamp,
					HighWaterMarkOffset: claim.HighWaterMarkOffset(),
				}
				h.consumer.parser.Messages(message)
			}); err != nil {
				h.wg.Done()
			}
			session.MarkMessage(msg, "")
		}
	}
}

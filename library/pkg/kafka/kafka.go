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

	client sarama.ConsumerGroup
	handle *handle

	ctx    context.Context
	cancel func()

	job *fanout.Fanout
	wg  *sync.WaitGroup
}

func (consumer *Consumer) Stop() {
	consumer.cancel()
	consumer.wg.Wait()
}

func NewConsumer(c *Config) (consumer *Consumer, err error) {
	version, err := sarama.ParseKafkaVersion(c.Kafka.Version)
	if err != nil {
		return nil, err
	}

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = version
	config.ClientID = fmt.Sprintf("%s-%s", c.Kafka.Name, c.Kafka.Group)

	switch strings.ToLower(c.Kafka.OffsetMode) {
	case "oldest", "earliest":
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	case "latest", "newest":
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	default:
	}

	switch strings.ToLower(c.Kafka.Rebalance) {
	case "range":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	case "sticky":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	case "roundrobin":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	default:
	}

	if c.Kafka.Sasl.Enable {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = c.Kafka.Sasl.User
		config.Net.SASL.Password = c.Kafka.Sasl.Password
		switch strings.ToUpper(c.Kafka.Sasl.Mechanism) {
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

	client, err := sarama.NewConsumerGroup(c.Kafka.Brokers, c.Kafka.Group, config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	consumer = &Consumer{
		config: config,
		client: client,
		handle: &handle{
			ready:       make(chan bool, 0),
			name:        c.Kafka.Name,
			logInterval: 30 * time.Second,
			wg:          &sync.WaitGroup{},
		},
		ctx:    ctx,
		cancel: cancel,
		wg:     &sync.WaitGroup{},
		job:    fanout.New("kafka", fanout.Worker(1), fanout.Buffer(10)),
	}

	consumer.handle.consumer = consumer
	consumer.wg.Add(1)
	go func() {
		defer consumer.wg.Done()
		for {
			select {
			case <-consumer.ctx.Done():
				log.Info("Terminating: context cancelled")
				return
			case err := <-consumer.client.Errors():
				log.Error("%s", err.Error())
			default:
				if err := consumer.client.Consume(consumer.ctx, c.Kafka.Topics, consumer.handle); err != nil {
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
					log.Info("Terminating:%v", ctx.Err())
					return
				}
				time.Sleep(time.Second)
				consumer.handle.ready = make(chan bool, 0)
			}
		}
	}()
	go func() {
		consumer.wg.Wait()
		if err := client.Close(); err != nil {
			log.Error("Error closing client: %v", err)
		}
	}()

	<-consumer.handle.ready
	log.Info("Kylin consumer up and running!...\n")
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
			log.Info("topic:%s, partition:%d session exit, waiting and process the left", claim.Topic(), claim.Partition())
			h.wg.Wait()
			return nil
		case <-logTick.C:
			log.Info("topic:%s, partition:%d", claim.Topic(), claim.Partition())
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			h.wg.Add(1)
			if err := h.consumer.job.DoWait(context.TODO(), func(ctx context.Context) {
				defer func() {
					h.wg.Done()
				}()
				//TODO: interface callback
				fmt.Printf("Message claimed: %s, timestamp:%v topic:%s, partition:%d, value:%s\n", h.name, msg.Timestamp, msg.Topic, msg.Partition, string(msg.Value))
			}); err != nil {
				h.wg.Done()
			}
		}
	}
}
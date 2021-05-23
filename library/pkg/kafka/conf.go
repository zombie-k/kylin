package kafka

import (
	"github.com/Shopify/sarama"
)

type Config struct {
	Consume *kafkaConf
	Job     struct{
		Worker int
		Buffer int
	}
}

type kafkaConf struct {
	Name    string
	Brokers []string
	Topics  []string
	Group   string

	Version    string
	OffsetMode string
	Rebalance  string

	Sasl struct {
		Enable    bool
		Mechanism string
		User      string
		Password  string
	}
}

func (c *Config) Builder() error {
	if c.Consume == nil {
		c.Consume = &kafkaConf{
			Name:       "kylin",
			Brokers:    []string{},
			Topics:     []string{},
			Group:      "",
			Version:    "0.10.2.1",
			OffsetMode: "latest",
			Rebalance:  "range",
		}
		c.Consume.Sasl.Enable = true
		c.Consume.Sasl.Mechanism = sarama.SASLTypePlaintext
		c.Consume.Sasl.User = ""
		c.Consume.Sasl.Password = ""
	}
	if _, err := sarama.ParseKafkaVersion(c.Consume.Version); err != nil {
		return err
	}
	if c.Job.Worker <= 0 {
		c.Job.Worker = 1
	}
	if c.Job.Buffer <= 0 {
		c.Job.Buffer = 1
	}
	return nil
}

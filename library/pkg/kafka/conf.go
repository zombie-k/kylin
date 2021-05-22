package kafka

import (
	"github.com/Shopify/sarama"
)

type Config struct {
	Kafka *kafkaConf
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

func (c *Config) Validate() error {
	if c.Kafka == nil {
		c.Kafka = &kafkaConf{
			Name:       "kylin",
			Brokers:    []string{},
			Topics:     []string{},
			Group:      "",
			Version:    "0.10.2.1",
			OffsetMode: "latest",
			Rebalance:  "range",
		}
		c.Kafka.Sasl.Enable = true
		c.Kafka.Sasl.Mechanism = sarama.SASLTypePlaintext
		c.Kafka.Sasl.User = ""
		c.Kafka.Sasl.Password = ""
	}
	if _, err := sarama.ParseKafkaVersion(c.Kafka.Version); err != nil {
		return err
	}
	return nil
}

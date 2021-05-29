package kafka

import (
	"errors"
	"github.com/Shopify/sarama"
	xtime "github.com/zombie-k/kylin/library/time"
	"strings"
)

type Config struct {
	Consume *kafkaConf
	Produce *ProduceConf
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

	Job struct {
		Worker int
		Buffer int
	}
}

type ProduceConf struct {
	Name      string
	Topic     string
	Brokers   []string
	Partition int
	Version   string

	// The duration the producer will wait to receive RequiredAcks.
	Timeout xtime.Duration
	// The partitioning scheme to use (hash, manual, random, roundrobin).
	Partitioner string
	// The compression method to use (none, gzip, snappy, lz4).
	Compression string
	// The name of the security protocol to talk to Kafka (PLAINTEXT, SSL) (default: PLAINTEXT).
	// The asyncProducer ID sent with every request to the brokers.
	ClientID string

	SecurityProtocol string
	// The path to a file that contains a set of root certificate authorities in PEM format
	// to trust when verifying broker certificates when SecurityProtocol=SSL
	// (leave empty to use the host's root CA set)."
	TlsRootCACerts string
	// The path to a file that contains the asyncProducer certificate to send to the broker
	// in PEM format if asyncProducer authentication is required when SecurityProtocol=SSL
	// (leave empty to disable asyncProducer authentication).
	TlsClientCert string
	// The path to a file that contains the asyncProducer private key linked to the asyncProducer certificate
	// in PEM format when SecurityProtocol=SSL (REQUIRED if TlsClientCert is provided).
	TlsClientKey string

	// The maximum number of unacknowledged requests the asyncProducer will send on a single connection
	// before blocking (default: 5).
	MaxOpenRequests int
	// The max permitted size of a message.
	MaxMessageBytes int
	// The required number of acks needed from the broker (all:WaitForAll, none:NoResponse, local:WaitForLoccal).
	RequiredAcks string

	// The best-effort frequency of flushes.
	FlushFrequency xtime.Duration
	// The best-effort number of bytes needed to trigger a flush.
	FlushBytes int
	// The best-effort number of messages needed to trigger a flush.
	FlushMessages int
	// The maximum number of messages the producer will send in a single request.
	FlushMaxMessages int

	// The number of events to buffer in internal and external channels.
	ChannelBufferSize int
	// The number of routines to send the messages from (sync only).
	Routines int
	// Turn on sarama logging to stderr
	Verbose bool
}

func (c *Config) Builder() error {
	if c.Consume == nil && c.Produce == nil {
		return errors.New("missing configuration consume or produce")
	}
	if _, err := sarama.ParseKafkaVersion(c.Consume.Version); err != nil {
		return err
	}
	if c.Consume.Job.Worker <= 0 {
		c.Consume.Job.Worker = 1
	}
	if c.Consume.Job.Buffer <= 0 {
		c.Consume.Job.Buffer = 1
	}

	if err := c.produceBuilder(); err != nil {
		return err
	}
	return nil
}

func (c *Config) produceBuilder() error {
	if _, err := sarama.ParseKafkaVersion(c.Produce.Version); err != nil {
		return err
	}
	if len(c.Produce.Brokers) == 0 {
		return errors.New("brokers is required")
	}
	if c.Produce.Topic == "" {
		return errors.New("topic is required")
	}

	if c.Produce.ClientID == "" {
		c.Produce.ClientID = "keylin"
	}

	switch strings.ToUpper(c.Produce.SecurityProtocol) {
	case "PLAINTEXT":
		c.Produce.SecurityProtocol = "PLAINTEXT"
	case "SSL":
		c.Produce.SecurityProtocol = "SSL"
	default:
		c.Produce.SecurityProtocol = "PLAINTEXT"
	}

	if c.Produce.MaxOpenRequests <= 0 {
		c.Produce.MaxOpenRequests = 5
	}

	if c.Produce.MaxMessageBytes <= 0 {
		c.Produce.MaxMessageBytes = 1000000
	}

	return nil
}

func (c *Config) parsePartitioner() sarama.PartitionerConstructor {
	switch c.Produce.Partitioner {
	case "manual":
		return sarama.NewManualPartitioner
	case "hash":
		return sarama.NewHashPartitioner
	case "random":
		return sarama.NewRandomPartitioner
	case "roundrobin":
		return sarama.NewRoundRobinPartitioner
	default:
		return sarama.NewRandomPartitioner
	}
}

func (c *Config) parseCompression() sarama.CompressionCodec {
	switch c.Produce.Compression {
	case "none":
		return sarama.CompressionNone
	case "gzip":
		return sarama.CompressionGZIP
	case "snappy":
		return sarama.CompressionSnappy
	case "lz4":
		return sarama.CompressionLZ4
	default:
		return sarama.CompressionNone
	}
}

func (c *Config) parseVersion(version string) sarama.KafkaVersion {
	ver, _ := sarama.ParseKafkaVersion(version)
	return ver
}

func (c *Config) parseRequiredAcks() sarama.RequiredAcks {
	switch c.Produce.RequiredAcks {
	case "none":
		return sarama.NoResponse
	case "local":
		return sarama.WaitForLocal
	case "all":
		return sarama.WaitForAll
	default:
		return sarama.WaitForLocal
	}
}

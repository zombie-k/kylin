package kafka

import (
	"context"
	"crypto/x509"
	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/tools/tls"
	"github.com/zombie-k/kylin/library/log"
	"io/ioutil"
	"sync"
	"time"
)

type Producer struct {
	c      *Config
	config *sarama.Config

	asyncProducer sarama.AsyncProducer
	client        sarama.Client
	loader        Loader

	ctx    context.Context
	cancel func()
	wg     sync.WaitGroup
}

func (producer *Producer) Stop() {
	producer.cancel()
	producer.wg.Wait()
}

func NewProducer(c *Config, loader Loader) (producer *Producer, err error) {
	config := sarama.NewConfig()

	config.ClientID = c.Produce.ClientID
	config.Version = c.parseVersion(c.Produce.Version)
	config.Net.MaxOpenRequests = c.Produce.MaxOpenRequests
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.MaxMessageBytes = c.Produce.MaxMessageBytes
	config.Producer.RequiredAcks = c.parseRequiredAcks()
	config.Producer.Partitioner = c.parsePartitioner()
	config.Producer.Compression = c.parseCompression()
	if c.Produce.ChannelBufferSize > 0 {
		config.ChannelBufferSize = c.Produce.ChannelBufferSize
	}
	if time.Duration(c.Produce.Timeout) > 0 {
		config.Producer.Timeout = time.Duration(c.Produce.Timeout)
	}
	if time.Duration(c.Produce.FlushFrequency) > 0 {
		config.Producer.Flush.Frequency = time.Duration(c.Produce.FlushFrequency)
	}
	if c.Produce.FlushBytes > 0 {
		config.Producer.Flush.Bytes = c.Produce.FlushBytes
	}
	if c.Produce.FlushMessages > 0 {
		config.Producer.Flush.Messages = c.Produce.FlushMessages
	}
	if c.Produce.FlushMaxMessages > 0 {
		config.Producer.Flush.MaxMessages = c.Produce.FlushMaxMessages
	}
	if c.Produce.SecurityProtocol == "SSL" {
		tlsConfig, err := tls.NewConfig(c.Produce.TlsClientCert, c.Produce.TlsClientKey)
		if err != nil {
			log.Error("failed to load asyncProducer certificate from: %s and private key from: %s: %v", c.Produce.TlsClientCert, c.Produce.TlsClientKey, err)
			return nil, err
		}
		if c.Produce.TlsRootCACerts != "" {
			rootCAsBytes, err := ioutil.ReadFile(c.Produce.TlsRootCACerts)
			if err != nil {
				log.Error("failed to read root CA certificates: %v", err)
				return nil, err
			}
			certPool := x509.NewCertPool()
			if !certPool.AppendCertsFromPEM(rootCAsBytes) {
				log.Error("failed to load root CA certificates from file: %s", c.Produce.TlsRootCACerts)
				return nil, err
			}
			tlsConfig.RootCAs = certPool
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}

	asyncProducer, err := sarama.NewAsyncProducer(c.Produce.Brokers, config)
	if err != nil {
		log.Error("failed to create producer: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	producer = &Producer{
		c:             c,
		config:        config,
		asyncProducer: asyncProducer,
		loader:        loader,
		ctx:           ctx,
		cancel:        cancel,
	}

	producer.wg.Add(1)
	go func() {
		defer producer.wg.Done()
		pwg := sync.WaitGroup{}
		for {
			select {
			case <-producer.ctx.Done():
				log.Debug("Terminating: context cancelled")
				pwg.Wait()
				return
			case msg := <-producer.asyncProducer.Errors():
				log.Error("Produce error: %+v. msg:%s", msg.Err, msg.Msg.Value)
			case msg := <-producer.asyncProducer.Successes():
				log.Debug("Message success! %d %s %d %d %s %s", msg.Timestamp, msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)
			case msg, ok := <-producer.loader.Inputs():
				if !ok {
					log.Debug("loader closing")
					pwg.Wait()
					return
				}
				pwg.Add(1)
				message := &sarama.ProducerMessage{
					Topic:     msg.Topic,
					Key:       sarama.ByteEncoder(msg.Key),
					Value:     sarama.ByteEncoder(msg.Value),
					Partition: msg.Partition,
				}
				producer.asyncProducer.Input() <- message
				pwg.Done()
			}
		}
	}()

	go func() {
		producer.wg.Wait()
		if err := producer.asyncProducer.Close(); err != nil {
			log.Error("Error closing asyncProducer: %v", err)
		}
	}()
	return
}

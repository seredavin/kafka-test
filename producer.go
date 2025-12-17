package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/IBM/sarama"
)

// KafkaProducer manages Kafka producer connection
type KafkaProducer struct {
	producer sarama.SyncProducer
	config   *Config
}

// NewKafkaProducer creates a new Kafka producer with mTLS support
func NewKafkaProducer(config *Config) (*KafkaProducer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Timeout = 10 * time.Second
	saramaConfig.Producer.Retry.Max = 3

	// Configure mTLS if enabled
	if config.UseAuth {
		tlsConfig, err := createTLSConfig(config.CertFile, config.KeyFile, config.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}

		saramaConfig.Net.TLS.Enable = true
		saramaConfig.Net.TLS.Config = tlsConfig
	}

	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &KafkaProducer{
		producer: producer,
		config:   config,
	}, nil
}

// createTLSConfig creates TLS configuration for mTLS
func createTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	// Load client certificate and key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	// Load CA certificate
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS12,
	}

	return tlsConfig, nil
}

// SendMessage sends a message to Kafka topic
func (p *KafkaProducer) SendMessage(key, value string) (partition int32, offset int64, err error) {
	msg := &sarama.ProducerMessage{
		Topic: p.config.Topic,
		Value: sarama.StringEncoder(value),
	}

	if key != "" {
		msg.Key = sarama.StringEncoder(key)
	}

	partition, offset, err = p.producer.SendMessage(msg)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to send message: %w", err)
	}

	return partition, offset, nil
}

// Close closes the producer connection
func (p *KafkaProducer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}

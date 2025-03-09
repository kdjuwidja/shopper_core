package util

import (
	"fmt"
	"os"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// KafkaProducerFactory creates and manages Kafka producers
type KafkaProducerFactory struct {
	config *kafka.ConfigMap
}

var (
	factory *KafkaProducerFactory
	once    sync.Once
)

// GetKafkaFactory returns the singleton factory instance
func GetKafkaFactory() (*KafkaProducerFactory, error) {
	var initErr error
	once.Do(func() {
		factory, initErr = initFactory()
	})
	if initErr != nil {
		return nil, fmt.Errorf("failed to initialize kafka factory: %v", initErr)
	}
	return factory, nil
}

// initFactory creates and configures a new factory instance
func initFactory() (*KafkaProducerFactory, error) {
	bootstrapServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	if bootstrapServers == "" {
		bootstrapServers = "kafka:29092" // fallback default
	}

	config := &kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"acks":              "all",    // Strongest delivery guarantee
		"retries":           3,        // Retry a few times before giving up
		"retry.backoff.ms":  100,      // Wait 100ms between retries
		"linger.ms":         5,        // Wait up to 5ms for batching
		"compression.type":  "snappy", // Use Snappy compression
	}

	return &KafkaProducerFactory{
		config: config,
	}, nil
}

// CreateKafkaProducer creates a new KafkaProducer instance
func (f *KafkaProducerFactory) CreateKafkaProducer() (*KafkaProducer, error) {
	producer, err := kafka.NewProducer(f.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %s", err)
	}

	return NewKafkaProducer(producer), nil
}

// CreateProducer creates a new Kafka producer instance
func (f *KafkaProducerFactory) CreateProducer() (*kafka.Producer, error) {
	producer, err := kafka.NewProducer(f.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %s", err)
	}
	return producer, nil
}

// CreateProducerWithDeliveryChannel creates a new producer with a delivery channel
func (f *KafkaProducerFactory) CreateProducerWithDeliveryChannel() (*kafka.Producer, chan kafka.Event, error) {
	producer, err := f.CreateProducer()
	if err != nil {
		return nil, nil, err
	}

	// Create delivery channel with reasonable buffer size
	deliveryChan := make(chan kafka.Event, 100)

	return producer, deliveryChan, nil
}

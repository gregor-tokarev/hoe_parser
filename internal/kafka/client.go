package kafka

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

// Client represents a Kafka client
type Client struct {
	config   *sarama.Config
	brokers  []string
	producer sarama.SyncProducer
	consumer sarama.Consumer
}

// NewClient creates a new Kafka client
func NewClient(brokers string) (*Client, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0 // Use a compatible Kafka version
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	brokerList := strings.Split(brokers, ",")

	client := &Client{
		config:  config,
		brokers: brokerList,
	}

	return client, nil
}

// InitProducer initializes the Kafka producer
func (c *Client) InitProducer() error {
	producer, err := sarama.NewSyncProducer(c.brokers, c.config)
	if err != nil {
		return err
	}
	c.producer = producer
	return nil
}

// InitConsumer initializes the Kafka consumer
func (c *Client) InitConsumer() error {
	consumer, err := sarama.NewConsumer(c.brokers, c.config)
	if err != nil {
		return err
	}
	c.consumer = consumer
	return nil
}

// SendMessage sends a message to a Kafka topic
func (c *Client) SendMessage(topic, key, value string) error {
	if c.producer == nil {
		return fmt.Errorf("producer not initialized")
	}

	message := &sarama.ProducerMessage{
		Topic:     topic,
		Key:       sarama.StringEncoder(key),
		Value:     sarama.StringEncoder(value),
		Timestamp: time.Now(),
	}

	partition, offset, err := c.producer.SendMessage(message)
	if err != nil {
		return err
	}

	log.Printf("Message sent to topic %s, partition %d, offset %d", topic, partition, offset)
	return nil
}

// ConsumeMessages consumes messages from a Kafka topic
func (c *Client) ConsumeMessages(ctx context.Context, topic string, handler func([]byte) error) error {
	if c.consumer == nil {
		return fmt.Errorf("consumer not initialized")
	}

	partitions, err := c.consumer.Partitions(topic)
	if err != nil {
		return err
	}

	for _, partition := range partitions {
		pc, err := c.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			return err
		}

		go func(pc sarama.PartitionConsumer) {
			defer pc.Close()
			for {
				select {
				case message := <-pc.Messages():
					if message != nil {
						if err := handler(message.Value); err != nil {
							log.Printf("Error handling message: %v", err)
						}
					}
				case err := <-pc.Errors():
					if err != nil {
						log.Printf("Consumer error: %v", err)
					}
				case <-ctx.Done():
					return
				}
			}
		}(pc)
	}

	return nil
}

// Close closes the Kafka client connections
func (c *Client) Close() error {
	var errors []error

	if c.producer != nil {
		if err := c.producer.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if c.consumer != nil {
		if err := c.consumer.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing client: %v", errors)
	}

	return nil
}

// HealthCheck checks if Kafka is accessible
func (c *Client) HealthCheck() error {
	client, err := sarama.NewClient(c.brokers, c.config)
	if err != nil {
		return err
	}
	defer client.Close()

	// Try to get broker list to verify connection
	brokers := client.Brokers()
	if len(brokers) == 0 {
		return fmt.Errorf("no brokers available")
	}

	return nil
}

package main

import (
	"fmt"
	"log"

	"github.com/gregor-tokarev/hoe_parser/internal/config"
	"github.com/gregor-tokarev/hoe_parser/internal/kafka"
)

func main() {
	cfg := config.Load()
	fmt.Printf("Loaded config: %+v\n", cfg)

	kafkaClient, err := kafka.NewClient(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka client: %v", err)
	}
	defer kafkaClient.Close()

	// Initialize producer for sending events (optional in development)
	if err := kafkaClient.InitProducer(); err != nil {
		log.Printf("Failed to initialize Kafka producer (this is normal if Kafka is not running): %v", err)
		log.Println("Continuing without Kafka functionality...")
	} else {
		// Test Kafka connection
		if err := kafkaClient.HealthCheck(); err != nil {
			log.Printf("Kafka health check failed: %v", err)
		} else {
			log.Println("Kafka connection verified successfully")
		}
	}
}

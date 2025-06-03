package config

import (
	"flag"
	"fmt"
	"strings"
)

// Config holds application configuration
type Config struct {
	Brokers       []string
	Topic         string
	ConsumerGroup string
	ProduceMode   bool
	ConsumeMode   bool
}

// Parse processes command line arguments and returns a Config
func Parse() (*Config, error) {
	var (
		brokers       string
		topic         string
		consumerGroup string
		produceMode   bool
		consumeMode   bool
	)

	flag.StringVar(&brokers, "b", "localhost:9092", "Kafka broker(s) separated by commas")
	flag.StringVar(&topic, "t", "", "Topic to produce to or consume from")
	flag.StringVar(&consumerGroup, "g", "kafkadog", "Consumer group ID (for consume mode)")
	flag.BoolVar(&produceMode, "p", false, "Producer mode - read from stdin and send to Kafka")
	flag.BoolVar(&consumeMode, "c", false, "Consumer mode - read from Kafka and write to stdout")

	flag.Parse()

	if topic == "" {
		return nil, fmt.Errorf("topic (-t) must be specified")
	}

	if !produceMode && !consumeMode {
		return nil, fmt.Errorf("either produce (-p) or consume (-c) mode must be specified")
	}

	if produceMode && consumeMode {
		return nil, fmt.Errorf("cannot use both produce (-p) and consume (-c) modes simultaneously")
	}

	return &Config{
		Brokers:       strings.Split(brokers, ","),
		Topic:         topic,
		ConsumerGroup: consumerGroup,
		ProduceMode:   produceMode,
		ConsumeMode:   consumeMode,
	}, nil
}

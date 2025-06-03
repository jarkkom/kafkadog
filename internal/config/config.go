package config

import (
	"flag"
	"fmt"
	"strings"

	ff "github.com/jarkkom/kafkadog/internal/format"
)

// Format represents the format for displaying/inputting messages
type Format string

// Config holds application configuration
type Config struct {
	Brokers       []string
	Topic         string
	ConsumerGroup string
	ProduceMode   bool
	ConsumeMode   bool
	Format        Format
}

// Parse processes command line arguments and returns a Config
func Parse() (*Config, error) {
	var (
		brokers       string
		topic         string
		consumerGroup string
		format        string
		produceMode   bool
		consumeMode   bool
	)

	availableFormats := ff.GetAvailableFormats()
	formatUsage := fmt.Sprintf("Message format: %s (for input in producer mode, output in consumer mode)", strings.Join(availableFormats, ", "))

	flag.StringVar(&brokers, "b", "localhost:9092", "Kafka broker(s) separated by commas")
	flag.StringVar(&topic, "t", "", "Topic to produce to or consume from")
	flag.StringVar(&consumerGroup, "G", "kafkadog", "Consumer group ID (for consume mode)")
	flag.StringVar(&format, "f", "raw", formatUsage)
	flag.BoolVar(&produceMode, "P", false, "Producer mode - read from stdin and send to Kafka")
	flag.BoolVar(&consumeMode, "C", false, "Consumer mode - read from Kafka and write to stdout")

	flag.Parse()

	if topic == "" {
		return nil, fmt.Errorf("topic (-t) must be specified")
	}

	// Default to consumer mode if neither mode is specified
	if !produceMode && !consumeMode {
		consumeMode = true
	}

	if produceMode && consumeMode {
		return nil, fmt.Errorf("cannot use both produce (-P) and consume (-C) modes simultaneously")
	}

	// Validate the format by checking if a codec exists for it
	_, err := ff.NewCodec(format)
	if err != nil {
		return nil, fmt.Errorf("invalid format (-f): %s. Must be one of: %s", format, strings.Join(availableFormats, ", "))
	}

	return &Config{
		Brokers:       strings.Split(brokers, ","),
		Topic:         topic,
		ConsumerGroup: consumerGroup,
		ProduceMode:   produceMode,
		ConsumeMode:   consumeMode,
		Format:        Format(format),
	}, nil
}

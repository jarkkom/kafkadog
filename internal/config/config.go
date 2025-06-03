package config

import (
	"flag"
	"fmt"
	"strings"
)

// Format represents the format for displaying/inputting messages
type Format string

const (
	// FormatRaw handles messages as-is
	FormatRaw Format = "raw"
	// FormatHex handles messages as hex encoded
	FormatHex Format = "hex"
	// FormatBase64 handles messages as base64 encoded
	FormatBase64 Format = "base64"
)

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

	flag.StringVar(&brokers, "b", "localhost:9092", "Kafka broker(s) separated by commas")
	flag.StringVar(&topic, "t", "", "Topic to produce to or consume from")
	flag.StringVar(&consumerGroup, "G", "kafkadog", "Consumer group ID (for consume mode)")
	flag.StringVar(&format, "f", "raw", "Message format: raw, hex, base64 (for input in producer mode, output in consumer mode)")
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

	// Validate format
	msgFormat := Format(format)
	if msgFormat != FormatRaw && msgFormat != FormatHex && msgFormat != FormatBase64 {
		return nil, fmt.Errorf("invalid format (-f): %s. Must be one of: raw, hex, base64", format)
	}

	return &Config{
		Brokers:       strings.Split(brokers, ","),
		Topic:         topic,
		ConsumerGroup: consumerGroup,
		ProduceMode:   produceMode,
		ConsumeMode:   consumeMode,
		Format:        msgFormat,
	}, nil
}

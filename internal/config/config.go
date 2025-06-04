package config

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	ff "github.com/jarkkom/kafkadog/internal/format"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Format represents the format for displaying/inputting messages
type Format string

// Config holds application configuration
type Config struct {
	Brokers        []string
	Topic          string
	ProduceMode    bool
	ConsumeMode    bool
	Format         Format
	DecodeProtobuf bool   // New field for protobuf decoding
	MessageCount   int    // Number of messages to read in consumer mode, 0 for unlimited
	ConsumerOffset string // Controls where to start consuming from: "beginning", "end", or an offset value
}

// Parse processes command line arguments and returns a Config
func Parse() (*Config, error) {
	var (
		brokers        string
		topic          string
		format         string
		produceMode    bool
		consumeMode    bool
		decodeProtobuf bool   // New variable for protobuf decoding flag
		messageCount   int    // Number of messages to read in consumer mode
		consumerOffset string // New variable for consumer offset flag
	)

	availableFormats := ff.GetAvailableFormats()
	formatUsage := fmt.Sprintf("Output format: %s", strings.Join(availableFormats, ", "))

	flag.StringVar(&brokers, "b", "localhost:9092", "Kafka broker(s) separated by commas")
	flag.StringVar(&topic, "t", "", "Topic to produce to or consume from")
	flag.StringVar(&format, "f", "raw", formatUsage)
	flag.BoolVar(&produceMode, "P", false, "Producer mode - read from stdin and send to Kafka")
	flag.BoolVar(&consumeMode, "C", false, "Consumer mode - read from Kafka and write to stdout")
	flag.BoolVar(&decodeProtobuf, "proto", false, "Decode binary data as Protocol Buffers before applying output format")
	flag.IntVar(&messageCount, "c", 0, "Number of messages to read in consumer mode (0 for unlimited)")
	flag.StringVar(&consumerOffset, "o", "beginning", "Consumer offset - where to start consuming from: 'beginning', 'end', or an offset value")

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
		Brokers:        strings.Split(brokers, ","),
		Topic:          topic,
		ProduceMode:    produceMode,
		ConsumeMode:    consumeMode,
		Format:         Format(format),
		DecodeProtobuf: decodeProtobuf,
		MessageCount:   messageCount,
		ConsumerOffset: consumerOffset,
	}, nil
}

// CreateConsumerOffset creates a kgo.Offset based on the ConsumerOffset config value
// It handles "beginning", "end", and numerical offset values (both absolute and relative)
func (c *Config) CreateConsumerOffset() (kgo.Offset, error) {
	// Default to consuming from the end of the topic
	kafkaOffset := kgo.NewOffset().AtEnd()

	// Handle empty string as end
	if c.ConsumerOffset == "" {
		return kafkaOffset, nil
	}

	switch c.ConsumerOffset {
	case "beginning":
		kafkaOffset = kgo.NewOffset().AtStart()
	case "end":
		kafkaOffset = kgo.NewOffset().AtEnd()
	default:
		// Try to parse as integer
		offsetVal, err := strconv.ParseInt(c.ConsumerOffset, 10, 64)
		if err != nil {
			return kafkaOffset, fmt.Errorf("invalid offset value '%s': %w", c.ConsumerOffset, err)
		}

		if offsetVal < 0 {
			// Negative value means relative offset from end
			kafkaOffset = kgo.NewOffset().Relative(offsetVal)
		} else {
			// Non-negative value is an absolute offset
			kafkaOffset = kgo.NewOffset().At(offsetVal)
		}
	}

	return kafkaOffset, nil
}

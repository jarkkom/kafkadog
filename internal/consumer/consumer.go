package consumer

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jarkkom/kafkadog/internal/config"
	"github.com/jarkkom/kafkadog/internal/format"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Consumer handles Kafka message consumption
type Consumer struct {
	client       *kgo.Client
	codec        format.Codec
	messageCount int // Number of messages to read, 0 for unlimited
}

// New creates a new Consumer instance
func New(client *kgo.Client, formatStr config.Format, messageCount int, protoImportDirs []string, messageType string) (*Consumer, error) {
	var codec format.Codec
	var err error

	// Handle protobuf format specially
	if string(formatStr) == "protobuf" {
		// Use schema-based protobuf codec if import dirs and message type are provided
		if len(protoImportDirs) > 0 && messageType != "" {
			codec, err = format.NewCodecWithSchema("protobuf", protoImportDirs, messageType)
		} else {
			codec, err = format.NewCodec("protobuf")
		}
		if err != nil {
			return nil, fmt.Errorf("failed to initialize protobuf codec: %w", err)
		}
	} else {
		codec, err = format.NewCodec(string(formatStr))
		if err != nil {
			return nil, err
		}
	}

	return &Consumer{
		client:       client,
		codec:        codec,
		messageCount: messageCount,
	}, nil
}

// Run starts the consumer writing to stdout
func (c *Consumer) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// Counter for messages read
	messagesRead := 0

	for {
		// Check if we've reached the message limit
		if c.messageCount > 0 && messagesRead >= c.messageCount {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			fetches := c.client.PollFetches(ctx)
			if fetches.IsClientClosed() {
				return
			}

			if errs := fetches.Errors(); len(errs) > 0 {
				for _, err := range errs {
					fmt.Fprintf(os.Stderr, "Error consuming from topic %s: %v\n", err.Topic, err.Err)
				}
				continue
			}

			fetches.EachRecord(func(record *kgo.Record) {
				// Check if we've reached the message limit
				if c.messageCount > 0 && messagesRead >= c.messageCount {
					return
				}

				// Apply the configured format (which handles protobuf if needed)
				encoded, err := c.codec.Encode(record.Value)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
					return
				}

				fmt.Println(string(encoded))

				// Increment the message counter
				messagesRead++
			})
		}
	}
}

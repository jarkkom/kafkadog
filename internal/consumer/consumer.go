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
	client         *kgo.Client
	codec          format.Codec
	decodeProtobuf bool
	protobufCodec  format.Codec // Separate codec for protobuf decoding
}

// New creates a new Consumer instance
func New(client *kgo.Client, formatStr config.Format, decodeProtobuf bool) (*Consumer, error) {
	codec, err := format.NewCodec(string(formatStr))
	if err != nil {
		return nil, err
	}

	var protoCodec format.Codec
	if decodeProtobuf {
		protoCodec, err = format.NewCodec("protobuf")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize protobuf decoder: %w", err)
		}
	}

	return &Consumer{
		client:         client,
		codec:          codec,
		decodeProtobuf: decodeProtobuf,
		protobufCodec:  protoCodec,
	}, nil
}

// Run starts the consumer writing to stdout
func (c *Consumer) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
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
				var value []byte
				var err error

				// First step: Optionally decode protobuf
				if c.decodeProtobuf {
					value, err = c.protobufCodec.Encode(record.Value)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error decoding protobuf: %v\n", err)
						return
					}
				} else {
					value = record.Value
				}

				// Second step: Apply the configured output format
				encoded, err := c.codec.Encode(value)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
					return
				}

				fmt.Println(string(encoded))
			})
		}
	}
}

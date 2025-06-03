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
	client *kgo.Client
	codec  format.Codec
}

// New creates a new Consumer instance
func New(client *kgo.Client, formatStr config.Format) (*Consumer, error) {
	codec, err := format.NewCodec(string(formatStr))
	if err != nil {
		return nil, err
	}

	return &Consumer{
		client: client,
		codec:  codec,
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
				encoded, err := c.codec.Encode(record.Value)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error encoding message: %v\n", err)
					return
				}
				fmt.Println(string(encoded))
			})
		}
	}
}

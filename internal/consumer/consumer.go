package consumer

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"sync"

	"github.com/jarkkom/kafkadog/internal/config"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Consumer handles Kafka message consumption
type Consumer struct {
	client *kgo.Client
	format config.Format
}

// New creates a new Consumer instance
func New(client *kgo.Client, format config.Format) *Consumer {
	return &Consumer{
		client: client,
		format: format,
	}
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
				var output string
				switch c.format {
				case config.FormatHex:
					output = hex.EncodeToString(record.Value)
				case config.FormatBase64:
					output = base64.StdEncoding.EncodeToString(record.Value)
				default: // FormatRaw
					output = string(record.Value)
				}
				fmt.Println(output)
			})
		}
	}
}

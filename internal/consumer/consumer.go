package consumer

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Consumer handles Kafka message consumption
type Consumer struct {
	client *kgo.Client
}

// New creates a new Consumer instance
func New(client *kgo.Client) *Consumer {
	return &Consumer{
		client: client,
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
				fmt.Println(string(record.Value))
			})
		}
	}
}

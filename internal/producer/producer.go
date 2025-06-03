package producer

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Producer handles Kafka message production
type Producer struct {
	client *kgo.Client
	topic  string
}

// New creates a new Producer instance
func New(client *kgo.Client, topic string) *Producer {
	return &Producer{
		client: client,
		topic:  topic,
	}
}

// Run starts the producer reading from stdin
func (p *Producer) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			record := &kgo.Record{
				Topic: p.topic,
				Value: scanner.Bytes(),
			}

			results := p.client.ProduceSync(ctx, record)
			if results.FirstErr() != nil {
				fmt.Fprintf(os.Stderr, "Failed to produce message: %v\n", results.FirstErr())
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
	}
}

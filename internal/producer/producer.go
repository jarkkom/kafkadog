package producer

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jarkkom/kafkadog/internal/config"
	"github.com/jarkkom/kafkadog/internal/format"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Producer handles Kafka message production
type Producer struct {
	client *kgo.Client
	topic  string
	codec  format.Codec
}

// New creates a new Producer instance
func New(client *kgo.Client, topic string, formatStr config.Format) (*Producer, error) {
	codec, err := format.NewCodec(string(formatStr))
	if err != nil {
		return nil, err
	}

	return &Producer{
		client: client,
		topic:  topic,
		codec:  codec,
	}, nil
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
			input := scanner.Bytes()
			value, err := p.codec.Decode(input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to decode input: %v\n", err)
				continue
			}

			record := &kgo.Record{
				Topic: p.topic,
				Value: value,
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

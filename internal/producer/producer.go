package producer

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"sync"

	"github.com/jarkkom/kafkadog/internal/config"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Producer handles Kafka message production
type Producer struct {
	client *kgo.Client
	topic  string
	format config.Format
}

// New creates a new Producer instance
func New(client *kgo.Client, topic string, format config.Format) *Producer {
	return &Producer{
		client: client,
		topic:  topic,
		format: format,
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
			input := scanner.Bytes()
			var value []byte
			var err error

			// Decode input based on format
			switch p.format {
			case config.FormatHex:
				value, err = hex.DecodeString(string(input))
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to decode hex input: %v\n", err)
					continue
				}
			case config.FormatBase64:
				value, err = base64.StdEncoding.DecodeString(string(input))
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to decode base64 input: %v\n", err)
					continue
				}
			default: // FormatRaw
				value = input
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

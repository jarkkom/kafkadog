package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/jarkkom/kafkadog/internal/config"
	"github.com/jarkkom/kafkadog/internal/consumer"
	"github.com/jarkkom/kafkadog/internal/producer"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
	}

	if cfg.ConsumeMode {
		// Get the consumer offset from configuration
		kafkaOffset, err := cfg.CreateConsumerOffset()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v, defaulting to end of topic\n", err)
			kafkaOffset = kgo.NewOffset().AtEnd()
		}

		// Set options for direct topic consumption without joining a consumer group
		opts = append(opts,
			kgo.ConsumeTopics(cfg.Topic),
			kgo.ConsumeResetOffset(kafkaOffset),
			kgo.ConsumerGroup(""), // Empty group ID prevents joining a consumer group
		)
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Kafka client: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	var wg sync.WaitGroup

	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "Received termination signal. Shutting down...")
		cancel()
	}()

	if cfg.ProduceMode {
		wg.Add(1)
		prod, err := producer.New(client, cfg.Topic, cfg.Format)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating producer: %v\n", err)
			os.Exit(1)
		}
		go prod.Run(ctx, &wg)
	}

	if cfg.ConsumeMode {
		wg.Add(1)
		cons, err := consumer.New(client, cfg.Format, cfg.MessageCount, cfg.ProtoImportDirs, cfg.MessageType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating consumer: %v\n", err)
			os.Exit(1)
		}
		go cons.Run(ctx, &wg)
	}

	wg.Wait()
}

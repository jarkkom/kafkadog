package config

import (
	"flag"
	"os"
	"testing"
)

// TestParse tests the Parse function with different command-line arguments
func TestParse(t *testing.T) {
	// Save original command line arguments and restore them after the test
	origArgs := os.Args
	defer func() {
		os.Args = origArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	tests := []struct {
		name          string
		args          []string
		expectedError bool
		check         func(t *testing.T, cfg *Config)
	}{
		{
			name:          "missing topic",
			args:          []string{"kafkadog", "-b", "localhost:9092"},
			expectedError: true,
			check:         nil,
		},
		{
			name:          "valid consumer config",
			args:          []string{"kafkadog", "-t", "test-topic"},
			expectedError: false,
			check: func(t *testing.T, cfg *Config) {
				if !cfg.ConsumeMode {
					t.Errorf("Expected ConsumeMode=true, got false")
				}
				if cfg.ProduceMode {
					t.Errorf("Expected ProduceMode=false, got true")
				}
				if cfg.Topic != "test-topic" {
					t.Errorf("Expected Topic='test-topic', got '%s'", cfg.Topic)
				}
				if len(cfg.Brokers) != 1 || cfg.Brokers[0] != "localhost:9092" {
					t.Errorf("Expected Brokers=['localhost:9092'], got %v", cfg.Brokers)
				}
				if string(cfg.Format) != "raw" {
					t.Errorf("Expected Format='raw', got '%s'", cfg.Format)
				}
			},
		},
		{
			name:          "valid producer config",
			args:          []string{"kafkadog", "-t", "test-topic", "-P"},
			expectedError: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.ConsumeMode {
					t.Errorf("Expected ConsumeMode=false, got true")
				}
				if !cfg.ProduceMode {
					t.Errorf("Expected ProduceMode=true, got false")
				}
			},
		},
		{
			name:          "conflicting modes",
			args:          []string{"kafkadog", "-t", "test-topic", "-P", "-C"},
			expectedError: true,
			check:         nil,
		},
		{
			name:          "custom format",
			args:          []string{"kafkadog", "-t", "test-topic", "-f", "hex"},
			expectedError: false,
			check: func(t *testing.T, cfg *Config) {
				if string(cfg.Format) != "hex" {
					t.Errorf("Expected Format='hex', got '%s'", cfg.Format)
				}
			},
		},
		{
			name:          "invalid format",
			args:          []string{"kafkadog", "-t", "test-topic", "-f", "invalid-format"},
			expectedError: true,
			check:         nil,
		},
		{
			name:          "message count",
			args:          []string{"kafkadog", "-t", "test-topic", "-c", "42"},
			expectedError: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.MessageCount != 42 {
					t.Errorf("Expected MessageCount=42, got %d", cfg.MessageCount)
				}
			},
		},
		{
			name:          "consumer offset beginning",
			args:          []string{"kafkadog", "-t", "test-topic", "-o", "beginning"},
			expectedError: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.ConsumerOffset != "beginning" {
					t.Errorf("Expected ConsumerOffset='beginning', got '%s'", cfg.ConsumerOffset)
				}
			},
		},
		{
			name:          "consumer offset end",
			args:          []string{"kafkadog", "-t", "test-topic", "-o", "end"},
			expectedError: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.ConsumerOffset != "end" {
					t.Errorf("Expected ConsumerOffset='end', got '%s'", cfg.ConsumerOffset)
				}
			},
		},
		{
			name:          "consumer offset numeric",
			args:          []string{"kafkadog", "-t", "test-topic", "-o", "42"},
			expectedError: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.ConsumerOffset != "42" {
					t.Errorf("Expected ConsumerOffset='42', got '%s'", cfg.ConsumerOffset)
				}
			},
		},
		{
			name:          "multiple brokers",
			args:          []string{"kafkadog", "-t", "test-topic", "-b", "broker1:9092,broker2:9092,broker3:9092"},
			expectedError: false,
			check: func(t *testing.T, cfg *Config) {
				expectedBrokers := []string{"broker1:9092", "broker2:9092", "broker3:9092"}
				if len(cfg.Brokers) != len(expectedBrokers) {
					t.Errorf("Expected %d brokers, got %d", len(expectedBrokers), len(cfg.Brokers))
				}
				for i, broker := range expectedBrokers {
					if i >= len(cfg.Brokers) || cfg.Brokers[i] != broker {
						t.Errorf("Expected broker '%s' at position %d, got '%s'", broker, i, cfg.Brokers[i])
					}
				}
			},
		},
		{
			name:          "protobuf decoding",
			args:          []string{"kafkadog", "-t", "test-topic", "-proto"},
			expectedError: false,
			check: func(t *testing.T, cfg *Config) {
				if !cfg.DecodeProtobuf {
					t.Errorf("Expected DecodeProtobuf=true, got false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			os.Args = tt.args

			cfg, err := Parse()
			
			// Check for expected errors
			if tt.expectedError && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// Run additional checks if no error occurred and check function is provided
			if !tt.expectedError && err == nil && tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

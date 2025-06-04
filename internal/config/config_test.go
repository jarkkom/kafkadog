package config

import (
	"strings"
	"testing"

	"github.com/twmb/franz-go/pkg/kgo"
)

func TestCreateConsumerOffset(t *testing.T) {
	tests := []struct {
		name           string
		consumerOffset string
		expectError    bool
		checkOffset    func(t *testing.T, offset kgo.Offset)
		errorContains  string // Expected substring in error message
	}{
		{
			name:           "beginning offset",
			consumerOffset: "beginning",
			expectError:    false,
			checkOffset: func(t *testing.T, offset kgo.Offset) {
				// The offset value is internal, so we can only check if the function
				// doesn't return an error for valid inputs
			},
			errorContains: "",
		},
		{
			name:           "end offset",
			consumerOffset: "end",
			expectError:    false,
			checkOffset: func(t *testing.T, offset kgo.Offset) {
				// The offset value is internal, so we can only check if the function
				// doesn't return an error for valid inputs
			},
			errorContains: "",
		},
		{
			name:           "absolute offset",
			consumerOffset: "42",
			expectError:    false,
			checkOffset: func(t *testing.T, offset kgo.Offset) {
				// The offset value is internal, so we can only check if the function
				// doesn't return an error for valid inputs
			},
			errorContains: "",
		},
		{
			name:           "relative offset",
			consumerOffset: "-10",
			expectError:    false,
			checkOffset: func(t *testing.T, offset kgo.Offset) {
				// The offset value is internal, so we can only check if the function
				// doesn't return an error for valid inputs
			},
			errorContains: "",
		},
		{
			name:           "invalid offset",
			consumerOffset: "invalid",
			expectError:    true,
			checkOffset:    nil,
			errorContains:  "invalid offset value",
		},
		{
			name:           "zero offset",
			consumerOffset: "0",
			expectError:    false,
			checkOffset: func(t *testing.T, offset kgo.Offset) {
				// The offset value is internal, so we can only check if the function
				// doesn't return an error for valid inputs
			},
			errorContains: "",
		},
		{
			name:           "empty offset",
			consumerOffset: "",
			expectError:    false,
			checkOffset: func(t *testing.T, offset kgo.Offset) {
				// Should default to AtEnd
			},
			errorContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				ConsumerOffset: tt.consumerOffset,
			}

			offset, err := cfg.CreateConsumerOffset()

			// Check if error was expected
			if tt.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// Check for expected error message if applicable
			if tt.expectError && err != nil && tt.errorContains != "" {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			}

			// If we have a check function and no error, run it
			if tt.checkOffset != nil && err == nil {
				tt.checkOffset(t, offset)
			}
		})
	}
}

// TestCreateConsumerOffsetDefaults tests various edge cases with consumer offsets
func TestCreateConsumerOffsetDefaults(t *testing.T) {
	t.Run("invalid offset returns error but defaults to end", func(t *testing.T) {
		cfg := &Config{
			ConsumerOffset: "invalid",
		}

		// We expect the function to return with an error and the default offset
		offset, err := cfg.CreateConsumerOffset()
		if err == nil {
			t.Errorf("Expected error for invalid offset, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "invalid offset value") {
			t.Errorf("Expected error containing 'invalid offset', got '%s'", err.Error())
		}

		// Can't directly check the offset value, but we can ensure the function completed
		_ = offset
	})

	t.Run("default is end when empty", func(t *testing.T) {
		cfg := &Config{
			ConsumerOffset: "",
		}

		// Empty offset should default to the default behavior without error
		_, err := cfg.CreateConsumerOffset()
		if err != nil {
			t.Errorf("Expected no error for empty offset, got %v", err)
		}
	})

	t.Run("very large numeric offset", func(t *testing.T) {
		cfg := &Config{
			ConsumerOffset: "9223372036854775807", // MaxInt64
		}

		_, err := cfg.CreateConsumerOffset()
		if err != nil {
			t.Errorf("Expected no error for large offset, got %v", err)
		}
	})

	t.Run("very negative offset", func(t *testing.T) {
		cfg := &Config{
			ConsumerOffset: "-9223372036854775808", // MinInt64
		}

		_, err := cfg.CreateConsumerOffset()
		if err != nil {
			t.Errorf("Expected no error for negative offset, got %v", err)
		}
	})
}

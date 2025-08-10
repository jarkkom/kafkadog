package format

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

// Codec is the interface for encoding/decoding messages in different formats
type Codec interface {
	// Decode transforms a byte array from the encoded format to raw bytes
	Decode(input []byte) ([]byte, error)
	// Encode transforms raw bytes to the encoded format
	Encode(input []byte) ([]byte, error)
}

// NewCodec returns a codec for the specified format
func NewCodec(format string) (Codec, error) {
	factory, ok := codecRegistry[format]
	if ok {
		return factory(), nil
	}

	return nil, fmt.Errorf("unsupported format: %s", format)
}

// NewCodecWithSchema returns a codec for the specified format with schema support
func NewCodecWithSchema(format string, importDirs []string, messageType string) (Codec, error) {
	// For schema-based protobuf, create a special codec
	if format == "protobuf" && len(importDirs) > 0 && messageType != "" {
		return NewProtoSchemaCodec(importDirs, messageType)
	}

	// Fall back to regular codec creation
	return NewCodec(format)
}

// Register built-in codecs
func init() {
	registerCodec("raw", func() Codec {
		return &RawCodec{}
	})

	registerCodec("hex", func() Codec {
		return &HexCodec{}
	})

	registerCodec("base64", func() Codec {
		return &Base64Codec{}
	})
}

// RawCodec handles messages without any encoding/decoding
type RawCodec struct{}

// Decode simply returns the input bytes as-is
func (c *RawCodec) Decode(input []byte) ([]byte, error) {
	return input, nil
}

// Encode simply returns the input bytes as-is
func (c *RawCodec) Encode(input []byte) ([]byte, error) {
	return input, nil
}

// HexCodec handles messages in hexadecimal format
type HexCodec struct{}

// Decode converts hex encoded bytes to raw bytes
func (c *HexCodec) Decode(input []byte) ([]byte, error) {
	// Trim any whitespace that might be present
	hexStr := strings.TrimSpace(string(input))
	output, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex: %w", err)
	}
	return output, nil
}

// Encode converts raw bytes to hex encoded bytes
func (c *HexCodec) Encode(input []byte) ([]byte, error) {
	return []byte(hex.EncodeToString(input)), nil
}

// Base64Codec handles messages in base64 format
type Base64Codec struct{}

// Decode converts base64 encoded bytes to raw bytes
func (c *Base64Codec) Decode(input []byte) ([]byte, error) {
	// Trim any whitespace that might be present
	b64Str := strings.TrimSpace(string(input))
	output, err := base64.StdEncoding.DecodeString(b64Str)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}
	return output, nil
}

// Encode converts raw bytes to base64 encoded bytes
func (c *Base64Codec) Encode(input []byte) ([]byte, error) {
	return []byte(base64.StdEncoding.EncodeToString(input)), nil
}

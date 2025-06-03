package format

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// Codec is the interface for encoding/decoding messages in different formats
type Codec interface {
	// Decode transforms a byte array from the encoded format to raw bytes
	Decode(input []byte) ([]byte, error)
	// Encode transforms raw bytes to the encoded format
	Encode(input []byte) ([]byte, error)
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
	output, err := hex.DecodeString(string(input))
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex: %w", err)
	}
	return output, nil
}

// Encode converts raw bytes to hex encoded bytes
func (c *HexCodec) Encode(input []byte) ([]byte, error) {
	output := make([]byte, hex.EncodedLen(len(input)))
	hex.Encode(output, input)
	return output, nil
}

// Base64Codec handles messages in base64 format
type Base64Codec struct{}

// Decode converts base64 encoded bytes to raw bytes
func (c *Base64Codec) Decode(input []byte) ([]byte, error) {
	output, err := base64.StdEncoding.DecodeString(string(input))
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}
	return output, nil
}

// Encode converts raw bytes to base64 encoded bytes
func (c *Base64Codec) Encode(input []byte) ([]byte, error) {
	output := make([]byte, base64.StdEncoding.EncodedLen(len(input)))
	base64.StdEncoding.Encode(output, input)
	return output, nil
}

// NewCodec returns a codec based on the specified format
func NewCodec(format string) (Codec, error) {
	switch format {
	case "raw":
		return &RawCodec{}, nil
	case "hex":
		return &HexCodec{}, nil
	case "base64":
		return &Base64Codec{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

package format

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"

	"google.golang.org/protobuf/encoding/protowire"
)

// ProtobufCodec handles protobuf binary format
type ProtobufCodec struct{}

// Decode is not implemented for protobuf codec (as producer)
func (c *ProtobufCodec) Decode(input []byte) ([]byte, error) {
	return nil, fmt.Errorf("protobuf encoding not implemented: use in consumer mode only")
}

// Encode transforms protobuf wire format to human-readable text
func (c *ProtobufCodec) Encode(input []byte) ([]byte, error) {
	var buf bytes.Buffer
	err := decodeProtobufMessage(&buf, input, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to decode protobuf: %w", err)
	}
	return buf.Bytes(), nil
}

// decodeProtobufMessage decodes a protobuf message into a human-readable format
func decodeProtobufMessage(w io.Writer, data []byte, indent int) error {
	indentStr := strings.Repeat("  ", indent)
	fmt.Fprintf(w, "{\n")

	// Map to store fields for ordered output
	fields := make(map[int]string)

	// Parse protobuf message
	b := data
	for len(b) > 0 {
		num, wireType, n := protowire.ConsumeTag(b)
		if n < 0 {
			return fmt.Errorf("failed to parse field number and wire type")
		}

		fieldNum := int(num)
		b = b[n:]

		var fieldStr string

		switch wireType {
		case protowire.VarintType:
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return fmt.Errorf("invalid varint value for field %d", fieldNum)
			}
			fieldStr = fmt.Sprintf("%s%d: %d", indentStr+"  ", fieldNum, v)
			b = b[n:]

		case protowire.Fixed32Type:
			v, n := protowire.ConsumeFixed32(b)
			if n < 0 {
				return fmt.Errorf("invalid fixed32 value for field %d", fieldNum)
			}
			fieldStr = fmt.Sprintf("%s%d: %v (32-bit)", indentStr+"  ", fieldNum, v)
			b = b[n:]

		case protowire.Fixed64Type:
			v, n := protowire.ConsumeFixed64(b)
			if n < 0 {
				return fmt.Errorf("invalid fixed64 value for field %d", fieldNum)
			}
			fieldStr = fmt.Sprintf("%s%d: %v (64-bit)", indentStr+"  ", fieldNum, v)
			b = b[n:]

		case protowire.BytesType:
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return fmt.Errorf("invalid bytes value for field %d", fieldNum)
			}

			// Try to decode as a nested message
			var nestedBuf bytes.Buffer
			if err := decodeProtobufMessage(&nestedBuf, v, indent+1); err == nil {
				fieldStr = fmt.Sprintf("%s%d: %s", indentStr+"  ", fieldNum, nestedBuf.String())
			} else {
				// Try to interpret as a string if it looks like ASCII text
				if isPrintableASCII(v) {
					fieldStr = fmt.Sprintf("%s%d: \"%s\"", indentStr+"  ", fieldNum, string(v))
				} else {
					// Fall back to showing the byte length and a hex preview
					preview := hex.EncodeToString(v)
					if len(preview) > 40 {
						preview = preview[:40] + "..."
					}
					fieldStr = fmt.Sprintf("%s%d: [%d bytes: %s]", indentStr+"  ", fieldNum, len(v), preview)
				}
			}
			b = b[n:]

		case protowire.StartGroupType:
			// For legacy support - we don't expect to encounter this often
			n := findEndGroup(b, num) // Fixed: Use the original protowire.Number type
			if n < 0 {
				return fmt.Errorf("invalid group value for field %d", fieldNum)
			}
			fieldStr = fmt.Sprintf("%s%d: [Group not decoded]", indentStr+"  ", fieldNum)
			b = b[n:]

		default:
			return fmt.Errorf("unsupported wire type: %d for field %d", wireType, fieldNum)
		}

		fields[fieldNum] = fieldStr
	}

	// Sort fields by field number for consistent output
	fieldNums := make([]int, 0, len(fields))
	for num := range fields {
		fieldNums = append(fieldNums, num)
	}
	sort.Ints(fieldNums)

	// Print fields in order
	for i, num := range fieldNums {
		fmt.Fprintf(w, "%s", fields[num])
		if i < len(fieldNums)-1 {
			fmt.Fprintf(w, ",\n")
		} else {
			fmt.Fprintf(w, "\n")
		}
	}

	fmt.Fprintf(w, "%s}", indentStr)
	return nil
}

// findEndGroup scans forward to find the end of a group
// Returns the number of bytes consumed if successful, or -1 if not found
func findEndGroup(b []byte, startFieldNum protowire.Number) int {
	origLen := len(b)
	for len(b) > 0 {
		num, wt, n := protowire.ConsumeTag(b)
		if n < 0 {
			return -1
		}
		b = b[n:]

		// Check if this is the end group marker we're looking for
		if wt == protowire.EndGroupType && num == startFieldNum {
			return origLen - len(b)
		}

		// Skip the value based on its wire type
		var valueLen int
		switch wt {
		case protowire.VarintType:
			_, valueLen = protowire.ConsumeVarint(b)
		case protowire.Fixed32Type:
			_, valueLen = protowire.ConsumeFixed32(b)
		case protowire.Fixed64Type:
			_, valueLen = protowire.ConsumeFixed64(b)
		case protowire.BytesType:
			_, valueLen = protowire.ConsumeBytes(b)
		case protowire.StartGroupType:
			// Recursively find the end of this nested group
			valueLen = findEndGroup(b, num)
		case protowire.EndGroupType:
			valueLen = 0 // We already consumed the tag
		}

		if valueLen < 0 {
			return -1
		}
		b = b[valueLen:]
	}

	return -1 // Didn't find the end group
}

// isPrintableASCII checks if a byte slice contains only printable ASCII characters
func isPrintableASCII(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	for _, b := range data {
		if b < 32 || b > 126 {
			return false
		}
	}
	return true
}

// Register the protobuf codec
func init() {
	registerCodec("protobuf", func() Codec {
		return &ProtobufCodec{}
	})
}

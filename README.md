# KafkaDog

## Important! Read before usage!

**This is a vibecoding exercise**.

**99% of the code, including this README is written using Copilot Agent mode, and no guarantees of correct behaviour are made.**

## Introduction

KafkaDog is a lightweight command-line utility for working with Apache Kafka. It allows you to easily produce and consume messages in various formats.

## Installation

```bash
go install github.com/jarkkom/kafkadog@latest
```

## Features

- Consume messages from Kafka topics
- Produce messages to Kafka topics
- Support for different output formats (raw, hex, base64)
- Protocol Buffer decoding support
- Simple and intuitive command-line interface

## Usage

### Basic Examples

#### Consuming Messages

By default, KafkaDog operates in consumer mode:

```bash
# Consume messages from a topic in raw format
kafkadog -t my-topic

# Consume messages from a specific broker
kafkadog -b kafka-broker:9092 -t my-topic

# Start consuming from the beginning of the topic
kafkadog -t my-topic -o beginning

# Start consuming from the end of the topic (only new messages)
kafkadog -t my-topic -o end

# Start consuming from a specific offset
kafkadog -t my-topic -o 100

# Start consuming from 10 messages before the current end
kafkadog -t my-topic -o -10
```

#### Producing Messages

Use the `-P` flag to switch to producer mode:

```bash
# Produce messages to a topic
echo "Hello Kafka" | kafkadog -t my-topic -P

# Send multiple messages (one per line)
cat messages.txt | kafkadog -t my-topic -P
```

### Format Options

The `-f` flag allows you to specify different formats for encoding/decoding messages:

```bash
# Consume messages and output in hex format
kafkadog -t my-topic -f hex

# Consume messages and output in base64 format
kafkadog -t my-topic -f base64

# Limit the number of consumed messages to 10
kafkadog -t my-topic -c 10

# Consume 5 messages in hex format
kafkadog -t my-topic -f hex -c 5

# Produce messages from hex input
echo "48656C6C6F204B61666B61" | kafkadog -t my-topic -P -f hex

# Produce messages from base64 input
echo "SGVsbG8gS2Fma2E=" | kafkadog -t my-topic -P -f base64
```

### Protocol Buffer Decoding

Use `-f protobuf` to decode binary messages as Protocol Buffers:

```bash
# Decode protobuf messages and display in wire format (human-readable)
kafkadog -t protobuf-topic -f protobuf

# Decode protobuf messages using schema files and output as JSON
kafkadog -t user-events -f protobuf -I ./proto -M User

# Use multiple import directories for schema-based decoding
kafkadog -t events -f protobuf -I ./proto,./schemas -M com.example.Event

# Decode with schema from specific directory structure
kafkadog -t orders -f protobuf -I ./api/proto -M shopping.Order
```

**Protocol Buffer Decoding Modes:**
- **Wire format decoding**: Use `-f protobuf` alone for human-readable field-by-field output
- **Schema-based decoding**: Use `-f protobuf` with `-I` (import directories) and `-M` (message type) for structured JSON output

**Note:** Schema-based decoding requires:
- `-I` flag with comma-separated directories containing .proto files
- `-M` flag with the message type name (e.g., `MessageName` or `package.MessageName`)

The schema decoder will attempt to compile all .proto files and ignore missing dependencies where possible.

### Combined Examples

```bash
# Consume messages from a remote broker, decode as protobuf wire format
kafkadog -b remote-kafka:9092 -t events -f protobuf

# Consume only 50 messages, decode as protobuf with schema-based JSON output
kafkadog -b remote-kafka:9092 -t binary-events -f protobuf -I ./schemas -M EventMessage -c 50

# Start consuming from offset 1000, decode as protobuf
kafkadog -b remote-kafka:9092 -t events -o 1000 -f protobuf

# Get the latest 20 messages from the topic
kafkadog -t my-topic -o -20 -c 20

# Produce binary data from hex representation
cat binary-data.hex | kafkadog -b kafka-broker:9092 -t binary-topic -P -f hex
```

## Command-Line Options

| Option | Description |
|--------|-------------|
| `-b` | Kafka broker(s) separated by commas (default: "localhost:9092") |
| `-t` | Topic to produce to or consume from (required) |
| `-f` | Format: raw, hex, base64, protobuf (default: "raw") |
| `-P` | Producer mode - read from stdin and send to Kafka |
| `-C` | Consumer mode - read from Kafka and write to stdout (default if neither -P nor -C specified) |
| `-c` | Number of messages to read in consumer mode (0 for unlimited, default: 0) |
| `-o` | Consumer offset - where to start consuming from: 'beginning', 'end', or an offset value (default: "end") |
| `-I` | Comma-separated directories to search for .proto files (for schema-based protobuf decoding) |
| `-M` | Message type for protobuf schema decoding (e.g., 'package.MessageName') |

## Examples with Actual Output

### Consuming Text Messages

```
$ kafkadog -t text-messages
Hello World!
This is a test message
Welcome to Kafka
```

### Consuming Binary Messages in Hex Format

```
$ kafkadog -t binary-data -f hex
48656c6c6f204b61666b61
a1b2c3d4e5f6
```

### Decoding Protocol Buffer Messages (Wire Format)

```
$ kafkadog -t user-events -f protobuf
{
  1: {
    1: "john.doe@example.com",
    2: "John Doe",
    3: 35
  },
  2: 1623841254,
  3: "login"
}
{
  1: {
    1: "jane.smith@example.com",
    2: "Jane Smith",
    3: 28
  },
  2: 1623842198,
  3: "profile_update"
}
```

### Schema-Based Protocol Buffer Decoding (JSON Output)

```
$ kafkadog -t user-events -f protobuf -I ./proto -M UserEvent
{
  "user": {
    "email": "john.doe@example.com",
    "name": "John Doe",
    "age": 35
  },
  "timestamp": "1623841254",
  "eventType": "login"
}
{
  "user": {
    "email": "jane.smith@example.com",
    "name": "Jane Smith",
    "age": 28
  },
  "timestamp": "1623842198",
  "eventType": "profile_update"
}
```

## License

Apache License 2.0

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

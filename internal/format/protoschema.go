package format

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// ProtoSchemaCodec handles protobuf decoding with schema files
type ProtoSchemaCodec struct {
	messageType protoreflect.MessageType
	importDirs  []string
	messageName string
}

// NewProtoSchemaCodec creates a new schema-based protobuf codec
func NewProtoSchemaCodec(importDirs []string, messageTypeName string) (*ProtoSchemaCodec, error) {
	if messageTypeName == "" {
		return nil, fmt.Errorf("message type name (-M) is required for schema-based protobuf decoding")
	}

	if len(importDirs) == 0 {
		return nil, fmt.Errorf("at least one import directory (-I) is required for schema-based protobuf decoding")
	}

	// Validate import directories exist
	for _, dir := range importDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return nil, fmt.Errorf("import directory does not exist: %s", dir)
		}
	}

	codec := &ProtoSchemaCodec{
		importDirs:  importDirs,
		messageName: messageTypeName,
	}

	// Load the schema and resolve the message type
	if err := codec.loadSchema(); err != nil {
		return nil, fmt.Errorf("failed to load schema: %w", err)
	}

	return codec, nil
}

// loadSchema compiles the proto files and resolves the message type
func (c *ProtoSchemaCodec) loadSchema() error {
	// Find all .proto files in the import directories
	protoFiles, err := c.findProtoFiles()
	if err != nil {
		return fmt.Errorf("failed to find proto files: %w", err)
	}

	if len(protoFiles) == 0 {
		return fmt.Errorf("no .proto files found in import directories: %v", c.importDirs)
	}

	// Create a compiler with the import directories
	compiler := protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{
			ImportPaths: c.importDirs,
		}),
	}

	// Compile the proto files
	files, err := compiler.Compile(context.Background(), protoFiles...)
	if err != nil {
		return fmt.Errorf("failed to compile proto files: %w", err)
	}

	// Create a file descriptor set
	var fileDescriptors []protoreflect.FileDescriptor
	for _, file := range files {
		fileDescriptors = append(fileDescriptors, file)
	}

	// Find the message type
	var messageDesc protoreflect.MessageDescriptor
	for _, fileDesc := range fileDescriptors {
		messageDesc = c.findMessageInFile(fileDesc, c.messageName)
		if messageDesc != nil {
			break
		}
	}

	if messageDesc == nil {
		return fmt.Errorf("message type '%s' not found in proto files", c.messageName)
	}

	// Create the message type
	c.messageType = dynamicpb.NewMessageType(messageDesc)

	return nil
}

// findProtoFiles recursively finds all .proto files in the import directories
func (c *ProtoSchemaCodec) findProtoFiles() ([]string, error) {
	var protoFiles []string

	for _, dir := range c.importDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".proto") {
				// Convert to relative path from the import directory
				relPath, err := filepath.Rel(dir, path)
				if err != nil {
					return err
				}
				protoFiles = append(protoFiles, relPath)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return protoFiles, nil
}

// findMessageInFile searches for a message type in a file descriptor
func (c *ProtoSchemaCodec) findMessageInFile(fileDesc protoreflect.FileDescriptor, messageName string) protoreflect.MessageDescriptor {
	// Try direct lookup first
	messageDesc := fileDesc.Messages().ByName(protoreflect.Name(messageName))
	if messageDesc != nil {
		return messageDesc
	}

	// Try with package prefix
	if fileDesc.Package() != "" {
		fullName := string(fileDesc.Package()) + "." + messageName
		if strings.HasSuffix(string(fileDesc.FullName()), fullName) {
			messages := fileDesc.Messages()
			for i := 0; i < messages.Len(); i++ {
				msg := messages.Get(i)
				if string(msg.FullName()) == fullName {
					return msg
				}
			}
		}
	}

	// Search nested messages recursively
	messages := fileDesc.Messages()
	for i := 0; i < messages.Len(); i++ {
		msg := messages.Get(i)
		if found := c.findNestedMessage(msg, messageName); found != nil {
			return found
		}
	}

	return nil
}

// findNestedMessage recursively searches for a message in nested messages
func (c *ProtoSchemaCodec) findNestedMessage(messageDesc protoreflect.MessageDescriptor, messageName string) protoreflect.MessageDescriptor {
	// Check if this message matches
	if string(messageDesc.Name()) == messageName || string(messageDesc.FullName()) == messageName {
		return messageDesc
	}

	// Search nested messages
	nested := messageDesc.Messages()
	for i := 0; i < nested.Len(); i++ {
		nestedMsg := nested.Get(i)
		if found := c.findNestedMessage(nestedMsg, messageName); found != nil {
			return found
		}
	}

	return nil
}

// Decode is not implemented for schema-based protobuf codec (producer mode not supported)
func (c *ProtoSchemaCodec) Decode(input []byte) ([]byte, error) {
	return nil, fmt.Errorf("schema-based protobuf encoding not implemented: use in consumer mode only")
}

// Encode transforms protobuf wire format to JSON using the schema
func (c *ProtoSchemaCodec) Encode(input []byte) ([]byte, error) {
	if c.messageType == nil {
		return nil, fmt.Errorf("message type not loaded")
	}

	// Create a new message instance
	message := c.messageType.New()

	// Unmarshal the protobuf data
	if err := proto.Unmarshal(input, message.Interface()); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf data: %w", err)
	}

	// Convert to JSON
	marshaler := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		UseProtoNames:   false,
		EmitUnpopulated: false,
	}

	jsonData, err := marshaler.Marshal(message.Interface())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	return jsonData, nil
}

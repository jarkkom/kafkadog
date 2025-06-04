.PHONY: build test clean vet fmt

# Binary name
BINARY=kafkadog
# Main package path
MAIN_PACKAGE=./cmd/kafkadog

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt

# Default target
all: build test

# Build the binary
build:
	@echo "Building $(BINARY)..."
	@$(GOBUILD) -o $(BINARY) $(MAIN_PACKAGE)

# Run tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v ./...

# No coverage targets needed per requirements

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -f $(BINARY)

# Run go vet
vet:
	@echo "Running go vet..."
	@$(GOVET) ./...

# Format code
fmt:
	@echo "Formatting code..."
	@$(GOFMT) ./...

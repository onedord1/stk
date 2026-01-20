.PHONY: build clean run test install

# Build variables
BINARY_NAME=systask
VERSION=1.0.0
BUILD_DIR=build
MAIN_PATH=./main.go

# Go build flags
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"

# Default target
all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: ./$(BINARY_NAME)"

# Build for all platforms
build-all: clean
	@mkdir -p $(BUILD_DIR)
	@echo "Building for Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "Building for Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@echo "Building for macOS AMD64..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@echo "Building for macOS ARM64..."
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "All builds complete in $(BUILD_DIR)/"

# Run the application
run: build
	@./$(BINARY_NAME)

# Run tests
test:
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@go test -cover ./...
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Install to GOPATH/bin
install:
	@go install $(LDFLAGS) $(MAIN_PATH)
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Cleaned build artifacts"

# Format code
fmt:
	@go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@golangci-lint run

# Tidy dependencies
tidy:
	@go mod tidy

# Download dependencies
deps:
	@go mod download

# Show help
help:
	@echo "SysTask Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build       - Build the application"
	@echo "  make build-all   - Build for all platforms"
	@echo "  make run         - Build and run"
	@echo "  make test        - Run tests"
	@echo "  make install     - Install to GOPATH/bin"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make fmt         - Format code"
	@echo "  make lint        - Lint code"
	@echo "  make tidy        - Tidy go.mod"
	@echo "  make deps        - Download dependencies"

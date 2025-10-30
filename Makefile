.PHONY: build build-cli build-all test clean fmt vet lint install-deps release-build

# Variables
BINARY_NAME=domain_converter
CLI_BINARY=domain_converter
CLI_PATH=./cmd/domain_converter
GO_FILES=$(shell find . -name '*.go' -type f)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

# Build platforms
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Default target
all: fmt vet test build-cli

# Build the CLI application
build-cli:
	go build $(LDFLAGS) -o $(CLI_BINARY) $(CLI_PATH)

# Build the plugin (library only)
build:
	go build -o $(BINARY_NAME) .

# Build for all platforms
build-all: clean
	@echo "Building for all platforms..."
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d/ -f1); \
		GOARCH=$$(echo $$platform | cut -d/ -f2); \
		OUTPUT_DIR=build/$$GOOS-$$GOARCH; \
		BINARY_NAME=$(CLI_BINARY); \
		if [ "$$GOOS" = "windows" ]; then \
			BINARY_NAME=$(CLI_BINARY).exe; \
		fi; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		mkdir -p $$OUTPUT_DIR; \
		GOOS=$$GOOS GOARCH=$$GOARCH CGO_ENABLED=0 go build $(LDFLAGS) -o $$OUTPUT_DIR/$$BINARY_NAME $(CLI_PATH); \
		if [ $$? -eq 0 ]; then \
			echo "✓ Built $$OUTPUT_DIR/$$BINARY_NAME"; \
		else \
			echo "✗ Failed to build for $$GOOS/$$GOARCH"; \
		fi; \
	done

# Create release archives
release-build: build-all
	@echo "Creating release archives..."
	@cd build && for dir in */; do \
		platform=$$(basename "$$dir"); \
		GOOS=$$(echo $$platform | cut -d- -f1); \
		if [ "$$GOOS" = "windows" ]; then \
			zip -r "../$(CLI_BINARY)_$(VERSION)_$$platform.zip" "$$dir"; \
		else \
			tar -czf "../$(CLI_BINARY)_$(VERSION)_$$platform.tar.gz" "$$dir"; \
		fi; \
		echo "✓ Created archive for $$platform"; \
	done

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run golangci-lint (if installed)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2"; \
	fi

# Install development dependencies
install-deps:
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	go clean
	rm -f $(CLI_BINARY)
	rm -f $(CLI_BINARY).exe
	rm -rf build/
	rm -f *.zip *.tar.gz
	rm -f coverage.out
	rm -f coverage.html

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Check for security issues (if gosec is installed)
security:
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Run: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Update dependencies
update-deps:
	go get -u ./...
	go mod tidy

# Run all checks (CI pipeline)
ci: fmt vet lint test security

# Development setup
dev-setup: install-deps
	@echo "Development environment setup complete!"
	@echo "Available commands:"
	@echo "  make test          - Run tests"
	@echo "  make test-coverage - Run tests with coverage"
	@echo "  make fmt           - Format code"
	@echo "  make lint          - Run linter"
	@echo "  make ci            - Run all CI checks"
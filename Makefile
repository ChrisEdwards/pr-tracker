# prt Makefile

# Version from git tag/commit
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

# Output directory
BIN_DIR = bin
BINARY = $(BIN_DIR)/prt

.PHONY: build test test-coverage clean install lint fmt help

## build: Build binary to bin/prt
build:
	@mkdir -p $(BIN_DIR)
	go build $(LDFLAGS) -o $(BINARY) ./cmd/prt

## test: Run all tests
test:
	go test ./...

## test-coverage: Run tests with coverage report
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## clean: Remove build artifacts
clean:
	rm -rf $(BIN_DIR) coverage.out coverage.html

## install: Install binary to /usr/local/bin
install: build
	cp $(BINARY) /usr/local/bin/prt

## lint: Run golangci-lint (if available)
lint:
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

## fmt: Format Go code
fmt:
	go fmt ./...

## help: Show this help
help:
	@echo "Available targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

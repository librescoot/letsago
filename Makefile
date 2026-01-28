BINARY_NAME := letsago
BUILD_DIR := bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-w -s -X main.version=$(VERSION)"

.PHONY: build build-arm build-host dist clean test fmt lint deps help

build:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

build-arm: build

build-host:
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

dist: build

clean:
	rm -rf $(BUILD_DIR)

test:
	go test -v ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

deps:
	go mod download
	go mod tidy

help:
	@echo "Targets:"
	@echo "  build      - Build for ARM (default)"
	@echo "  build-arm  - Alias for build"
	@echo "  build-host - Build for current host"
	@echo "  dist       - Stripped ARM binary"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  fmt        - Format code"
	@echo "  lint       - Run linter"
	@echo "  deps       - Download and tidy dependencies"

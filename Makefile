# Makefile for letsago - Vehicle State Watcher

# Binary name
BINARY_NAME=letsago

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Build flags
# -s -w: strip debugging information
# -extldflags: pass flags to external linker
# -trimpath: remove filesystem paths from resulting binary for reproducible builds
LDFLAGS=-ldflags="-s -w -extldflags '-static'"
GCFLAGS=-gcflags="-N -l"

# Set arm compiler flags
ARM_GOOS=linux
ARM_GOARCH=arm
ARM_GOARM=7

# Source files
SRC_FILES=main.go

.PHONY: all build clean test help dist-arm

# Default build for host system
all: build

# Build for the host system
build:
	@echo "Building for host system..."
	$(GOBUILD) -o $(BINARY_NAME) $(SRC_FILES)
	@echo "Build completed: $(BINARY_NAME)"

# Build optimized binary for ARMv7l
dist-arm:
	@echo "Building optimized ARMv7l binary..."
	GOOS=$(ARM_GOOS) GOARCH=$(ARM_GOARCH) GOARM=$(ARM_GOARM) $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-arm $(SRC_FILES)
	@echo "Stripping binary..."
	arm-linux-gnueabihf-strip $(BINARY_NAME)-arm 2>/dev/null || echo "Warning: arm-linux-gnueabihf-strip not found, skipping strip step"
	@echo "Build completed: $(BINARY_NAME)-arm"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME) $(BINARY_NAME)-arm

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Help information
help:
	@echo "Make targets:"
	@echo "  all        - Default target, same as build"
	@echo "  build      - Build for host system"
	@echo "  dist-arm   - Build optimized binary for ARMv7l"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  help       - Show this help message"

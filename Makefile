.PHONY: build build-release test lint clean install deps package-deb package-rpm package help

# Variables
BINARY_NAME := llm-imager
CMD_PATH := ./cmd/llm-imager
DIST_DIR := dist
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GOARCH ?= $(shell go env GOARCH)
GOOS ?= $(shell go env GOOS)
LDFLAGS := -s -w -X main.version=$(VERSION)

# Default target
all: build

# Build binary for current platform
build:
	go build -o $(BINARY_NAME) $(CMD_PATH)

# Build optimized binary for release
build-release:
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME) $(CMD_PATH)

# Build for specific platform (usage: make build-cross GOOS=linux GOARCH=arm64)
build-cross:
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="$(LDFLAGS)" \
		-o $(DIST_DIR)/$(BINARY_NAME)_$(GOOS)_$(GOARCH) $(CMD_PATH)

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html

# Install binary to GOPATH/bin
install:
	go install $(CMD_PATH)

# Install to system (requires root)
install-system: build-release
	install -Dm755 $(DIST_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

# Build DEB package
package-deb: build-release
	@command -v nfpm >/dev/null 2>&1 || { echo "nfpm not found. Install: go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest"; exit 1; }
	VERSION=$(VERSION) GOARCH=$(GOARCH) envsubst < nfpm.yaml > nfpm-resolved.yaml
	nfpm package -f nfpm-resolved.yaml -p deb -t $(DIST_DIR)/
	rm -f nfpm-resolved.yaml

# Build RPM package
package-rpm: build-release
	@command -v nfpm >/dev/null 2>&1 || { echo "nfpm not found. Install: go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest"; exit 1; }
	VERSION=$(VERSION) GOARCH=$(GOARCH) envsubst < nfpm.yaml > nfpm-resolved.yaml
	nfpm package -f nfpm-resolved.yaml -p rpm -t $(DIST_DIR)/
	rm -f nfpm-resolved.yaml

# Build all packages
package: package-deb package-rpm

# Run the application
run: build
	./$(BINARY_NAME)

# Show help
help:
	@echo "llm-imager Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build          Build binary for current platform"
	@echo "  build-release  Build optimized binary for release"
	@echo "  build-cross    Build for specific platform (GOOS=linux GOARCH=arm64)"
	@echo "  deps           Download and tidy dependencies"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  lint           Run golangci-lint"
	@echo "  clean          Remove build artifacts"
	@echo "  install        Install to GOPATH/bin"
	@echo "  install-system Install to /usr/local/bin (requires root)"
	@echo "  package-deb    Build DEB package"
	@echo "  package-rpm    Build RPM package"
	@echo "  package        Build all packages (DEB + RPM)"
	@echo "  run            Build and run the application"
	@echo "  help           Show this help"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION        Version string (default: git tag or 'dev')"
	@echo "  GOOS           Target OS for cross-compilation"
	@echo "  GOARCH         Target architecture for cross-compilation"

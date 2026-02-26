.PHONY: build clean test test-unit test-integration install dev-install deps fmt lint man docs-gen release-snapshot release check help

BINARY_NAME=opsgenie-cli
VERSION=$(shell git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo dev)
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/
	go clean

# Smoke tests - no API key needed
test: build
	@BINARY=./$(BINARY_NAME) bash smoke_test.sh

# Unit tests
test-unit:
	@echo "=== $(BINARY_NAME) unit tests ==="
	go test -v -race -coverprofile=coverage.out ./pkg/...
	@go tool cover -func=coverage.out | tail -1

# Integration tests - requires API key
test-integration: build
	@echo "=== $(BINARY_NAME) integration tests ==="
	@if [ -z "$${OPSGENIE_API_KEY:-}" ]; then \
		echo "SKIP: OPSGENIE_API_KEY not set"; exit 0; \
	fi
	go test -v -run TestIntegration -count=1 -timeout 10m ./...

install: build
	sudo install -m 755 $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

dev-install: build
	sudo ln -sf $(PWD)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

deps:
	go mod download
	go mod tidy

fmt:
	go fmt ./...

lint:
	golangci-lint run

# Release via GoReleaser (local snapshot for testing)
release-snapshot:
	goreleaser release --snapshot --clean

# Full GoReleaser release
release:
	goreleaser release --clean

check: fmt lint test test-unit

# Generate man pages via cobra/doc
man:
	@echo "Generating man pages..."
	@mkdir -p docs/man
	go run ./cmd/gendocs man docs/man
	@echo "Man pages generated in docs/man/"

# Generate markdown CLI reference via cobra/doc
docs-gen:
	@echo "Generating markdown docs..."
	@mkdir -p docs/cli
	go run ./cmd/gendocs markdown docs/cli
	@echo "CLI docs generated in docs/cli/"

help:
	@echo "Available targets:"
	@echo "  build            - Build the binary"
	@echo "  clean            - Clean build artifacts"
	@echo "  test             - Run smoke tests (no API key needed)"
	@echo "  test-unit        - Run unit tests with coverage"
	@echo "  test-integration - Run integration tests (requires OPSGENIE_API_KEY)"
	@echo "  install          - Install to /usr/local/bin/"
	@echo "  dev-install      - Symlink for development"
	@echo "  deps             - Download and tidy dependencies"
	@echo "  fmt              - Format code"
	@echo "  lint             - Lint with golangci-lint"
	@echo "  release-snapshot - Local GoReleaser test build"
	@echo "  release          - Full GoReleaser release"
	@echo "  man              - Generate man pages via cobra/doc"
	@echo "  docs-gen         - Generate markdown CLI reference via cobra/doc"
	@echo "  check            - Run fmt, lint, test, test-unit"
	@echo "  help             - Show this help"

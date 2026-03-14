BINARY_NAME=portyanka
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"

.PHONY: build test lint clean install release help

help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  test           - Run tests"
	@echo "  lint           - Run linter"
	@echo "  clean          - Remove build artifacts"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  release        - Build release binaries for all platforms"

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo ""
	@ls -lh $(BINARY_NAME)
	@echo "Binary: ./$(BINARY_NAME)"

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	rm -rf dist/

install:
	go install $(LDFLAGS) .

release:
	@echo "Building release binaries..."
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -trimpath -o dist/$(BINARY_NAME)-linux-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -trimpath -o dist/$(BINARY_NAME)-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -trimpath -o dist/$(BINARY_NAME)-darwin-arm64 .
	@echo ""
	@ls -lh dist/
	@echo ""
	@echo "Release binaries created in dist/"

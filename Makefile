# Hyperliquid Go SDK Makefile

.PHONY: build test clean fmt vet lint install-deps

# Build the project
build:
	go build ./...

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	go clean ./...
	rm -f coverage.out

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Install development dependencies
install-deps:
	go mod download
	go mod tidy

# Run all checks
check: fmt vet test

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the project"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  install-deps  - Install dependencies"
	@echo "  check         - Run fmt, vet, and test"
	@echo "  help          - Show this help message"

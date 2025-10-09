.PHONY: test coverage lint build clean help

# Default target
help:
	@echo "Shape Parser Library - Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  test      - Run all tests with race detection"
	@echo "  coverage  - Generate test coverage report"
	@echo "  lint      - Run golangci-lint"
	@echo "  build     - Build all packages"
	@echo "  clean     - Remove build artifacts and coverage files"
	@echo "  help      - Show this help message"

# Run tests with race detection
test:
	go test -v -race ./...

# Generate coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linters
lint:
	golangci-lint run

# Build all packages
build:
	go build ./...

# Clean build artifacts
clean:
	rm -f coverage.out coverage.html
	go clean ./...

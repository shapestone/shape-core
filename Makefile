.PHONY: test coverage lint build clean help bench bench-compare test-short test-verbose integration security-scan deps-verify deps-update fmt vet

# Default target
help:
	@echo "Shape Parser Library - Makefile"
	@echo ""
	@echo "Testing:"
	@echo "  test           - Run all tests with race detection"
	@echo "  test-short     - Run tests with -short flag for quick feedback"
	@echo "  test-verbose   - Run tests with extra verbosity"
	@echo "  integration    - Run integration tests only"
	@echo "  coverage       - Generate test coverage report"
	@echo ""
	@echo "Performance:"
	@echo "  bench          - Run all benchmarks"
	@echo "  bench-compare  - Compare benchmark results (requires old.bench file)"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint           - Run golangci-lint"
	@echo "  fmt            - Format code with gofmt"
	@echo "  vet            - Run go vet"
	@echo "  security-scan  - Run security scanners (gosec)"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps-verify    - Verify and tidy dependencies"
	@echo "  deps-update    - Update dependencies interactively"
	@echo ""
	@echo "Build:"
	@echo "  build          - Build all packages"
	@echo "  clean          - Remove build artifacts and coverage files"
	@echo ""
	@echo "  help           - Show this help message"

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
	golangci-lint run --timeout=5m

# Build all packages
build:
	go build ./...

# Clean build artifacts
clean:
	rm -f coverage.out coverage.html *.bench
	go clean ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem -run=^$$ ./...

# Compare benchmark results
bench-compare:
	@if [ ! -f old.bench ]; then \
		echo "Error: old.bench file not found"; \
		echo "Run: make bench > old.bench"; \
		echo "Then make changes and run: make bench > new.bench"; \
		echo "Finally run: make bench-compare"; \
		exit 1; \
	fi
	@if [ ! -f new.bench ]; then \
		echo "Error: new.bench file not found"; \
		echo "Run: make bench > new.bench"; \
		exit 1; \
	fi
	@command -v benchcmp >/dev/null 2>&1 || { \
		echo "Installing benchcmp..."; \
		go install golang.org/x/tools/cmd/benchcmp@latest; \
	}
	benchcmp old.bench new.bench

# Run tests with -short flag
test-short:
	go test -short ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run integration tests
integration:
	go test -v -tags=integration ./...

# Run security scanner
security-scan:
	@command -v gosec >/dev/null 2>&1 || { \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	}
	gosec -fmt=json -out=gosec-report.json ./...
	@echo "Security scan complete. Report: gosec-report.json"

# Verify and tidy dependencies
deps-verify:
	go mod verify
	go mod tidy
	@git diff --exit-code go.mod go.sum || { \
		echo "go.mod or go.sum has uncommitted changes"; \
		exit 1; \
	}

# Update dependencies
deps-update:
	go get -u ./...
	go mod tidy
	@echo "Dependencies updated. Review changes and run tests."

# Format code
fmt:
	gofmt -s -w .
	@echo "Code formatted"

# Run go vet
vet:
	go vet ./...

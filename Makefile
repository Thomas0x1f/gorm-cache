.PHONY: test test-coverage test-race fmt vet lint clean examples

# Test all packages
test:
	@echo "Running tests..."
	go test -v ./...

# Test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Test with race detector
test-race:
	@echo "Running tests with race detector..."
	go test -race ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	golangci-lint run ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f coverage.out coverage.html
	find . -name "*.db" -type f -delete
	go clean ./...

# Run all examples
examples:
	@echo "Running basic example..."
	cd examples/basic && go run main.go
	@echo "\nRunning context_usage example..."
	cd examples/context_usage && go run main.go
	@echo "\nRunning skip_cache example..."
	cd examples/skip_cache && go run main.go
	@echo "\nRunning model_selection example..."
	cd examples/model_selection && go run main.go

# Run example with Redis (requires Redis running)
example-redis:
	@echo "Running Redis example (make sure Redis is running)..."
	cd examples/redis && go run main.go

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Build examples
build-examples:
	@echo "Building examples..."
	cd examples/basic && go build -o ../../bin/basic
	cd examples/context_usage && go build -o ../../bin/context_usage
	cd examples/redis && go build -o ../../bin/redis
	cd examples/skip_cache && go build -o ../../bin/skip_cache
	cd examples/model_selection && go build -o ../../bin/model_selection

# Run all checks
check: fmt vet test
	@echo "All checks passed!"

# Help
help:
	@echo "Available targets:"
	@echo "  test              - Run all tests"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  test-race         - Run tests with race detector"
	@echo "  fmt               - Format code"
	@echo "  vet               - Vet code"
	@echo "  lint              - Lint code (requires golangci-lint)"
	@echo "  clean             - Clean build artifacts"
	@echo "  examples          - Run all examples"
	@echo "  example-redis     - Run Redis example"
	@echo "  deps              - Install dependencies"
	@echo "  build-examples    - Build all examples"
	@echo "  check             - Run all checks (fmt, vet, test)"
	@echo "  help              - Show this help message"

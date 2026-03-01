.PHONY: build run test test-coverage lint clean help

# Build the application
build:
	@echo "Building server..."
	go build -o bin/server cmd/server/main.go
	@echo "Build complete! Binary at bin/server"

# Run the application
run:
	@echo "Running server..."
	@echo "Starting Flight Search API on http://localhost:8080"
	@echo "API endpoints:"
	@echo "  - POST   /api/v1/search"
	@echo "  - POST   /api/v1/search/filter"
	@echo "  - GET    /api/v1/cache/stats"
	@echo "  - DELETE /api/v1/cache"
	@echo "  - GET    /api/v1/providers"
	@echo "  - GET    /api/v1/health"
	@echo ""
	go run cmd/server/main.go

# Run all tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

# Show test coverage in browser
test-coverage:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out
	@echo "Clean complete!"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Show help
help:
	@echo "Available commands:"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the application"
	@echo "  make test          - Run all tests with race detection"
	@echo "  make test-coverage - Generate and view test coverage"
	@echo "  make lint          - Run linter"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make deps          - Download and tidy dependencies"
	@echo "  make help          - Show this help message"

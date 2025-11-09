.PHONY: test bench cover lint fmt build clean install help quick-bench race

# Default target
.DEFAULT_GOAL := help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
BINARY_NAME=duplicatecheck

help: ## Display this help message
	@echo "DuplicateCheck - Makefile commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

test: ## Run all tests with race detector
	@echo "Running tests with race detector..."
	$(GOTEST) -v -race -timeout 30s ./...

test-short: ## Run tests without race detector (faster)
	@echo "Running tests..."
	$(GOTEST) -v ./...

race: ## Run tests with race detector and verbose output
	@echo "Running race condition tests..."
	$(GOTEST) -race -v ./...

bench: ## Run all benchmarks
	@echo "Running all benchmarks..."
	$(GOTEST) -bench=. -benchmem -benchtime=3s

quick-bench: ## Run quick performance matrix (recommended)
	@echo "Running quick performance matrix..."
	$(GOTEST) -bench=BenchmarkQuickMatrix -benchtime=1x -timeout=10m

critical-bench: ## Run critical benchmarks (N-grams, Hybrid vs Naive)
	@echo "Running critical benchmarks..."
	$(GOTEST) -bench="BenchmarkGetNgrams|BenchmarkHybridVsNaive" -benchmem -benchtime=3s

cover: ## Generate and view test coverage report
	@echo "Generating coverage report..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Opening in browser..."
	@open coverage.html 2>/dev/null || xdg-open coverage.html 2>/dev/null || echo "Please open coverage.html manually"

cover-text: ## Display test coverage in terminal
	@echo "Test coverage:"
	$(GOTEST) -cover ./...

lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: brew install golangci-lint" && exit 1)
	golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) ./...
	@which goimports > /dev/null && goimports -w . || echo "Tip: Install goimports with: go install golang.org/x/tools/cmd/goimports@latest"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...

tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	$(GOMOD) tidy

build: ## Build the binary
	@echo "Building..."
	$(GOBUILD) -o bin/$(BINARY_NAME) -v

clean: ## Clean build artifacts and test cache
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "Clean complete"

install: ## Install the package
	@echo "Installing..."
	$(GOCMD) install

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download

verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	$(GOMOD) verify

all: fmt vet test build ## Run fmt, vet, test, and build
	@echo "All checks passed!"

ci: fmt vet lint race cover-text ## Run all CI checks (fmt, vet, lint, race, cover)
	@echo "CI checks complete!"

check: fmt vet test-short ## Quick check (fmt, vet, test-short)
	@echo "Quick check passed!"

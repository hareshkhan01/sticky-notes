# Stick build targets. `make help` lists them.

BINARY  := stick
MODULE  := $(shell go list -m)
VERSION := $(shell git describe --tags --always 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: help build test vet fmt run clean install

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'

build: ## Build the stick binary
	go build -ldflags '$(LDFLAGS)' -o $(BINARY) .

test: ## Run all tests
	go test ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format all Go files
	gofmt -w .

run: build ## Build and show help
	./$(BINARY) --help

install: build ## Install stick into $GOPATH/bin
	install -m 0755 $(BINARY) $(shell go env GOPATH)/bin/$(BINARY)

clean: ## Remove build artifacts
	rm -f $(BINARY)

BINARY := agent-statsig
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/agent-statsig

test:
	go test ./... -count=1

test-short:
	go test ./... -count=1 -short

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
	goimports -w .

clean:
	rm -f $(BINARY)
	rm -f release/agent-statsig-*

dev:
	go run ./cmd/agent-statsig $(ARGS)

vet:
	go vet ./...

.PHONY: build test test-short lint fmt clean dev vet

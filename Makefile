VERSION ?= dev
LDFLAGS := -X github.com/revazi/tasklight/internal/cli.Version=$(VERSION)

.PHONY: test vet build run clean

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -ldflags "$(LDFLAGS)" -o bin/tasklight ./cmd/tasklight

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/tasklight --help

clean:
	rm -rf bin coverage.out

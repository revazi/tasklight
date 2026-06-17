.PHONY: test vet build run

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -o bin/tasklight ./cmd/tasklight

run:
	go run ./cmd/tasklight --help

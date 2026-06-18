VERSION ?= dev
LDFLAGS := -X github.com/revazi/tasklight/internal/cli.Version=$(VERSION)

.PHONY: test vet build run npm-package clean

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -ldflags "$(LDFLAGS)" -o bin/tasklight ./cmd/tasklight

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/tasklight --help

npm-package:
	npm --prefix npm/tasklight-cli run build:vendor
	npm --prefix npm/tasklight-cli run test:local
	npm --prefix npm/tasklight-cli run pack:check

clean:
	rm -rf bin coverage.out npm/tasklight-cli/vendor npm/tasklight-cli/*.tgz

VERSION ?= dev
LDFLAGS := -X github.com/revazi/tasklight/internal/cli.Version=$(VERSION)

.PHONY: test vet build run macos-helper npm-package clean

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -ldflags "$(LDFLAGS)" -o bin/tasklight ./cmd/tasklight

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/tasklight --help

macos-helper:
	./scripts/build-macos-helper.sh bin darwin-arm64

npm-package:
	npm --prefix npm/tasklight-cli run build:vendor
	npm --prefix npm/tasklight-cli run test:local
	npm --prefix npm/tasklight-cli run pack:check

clean:
	rm -rf bin coverage.out npm/tasklight-cli/vendor npm/tasklight-cli/*.tgz

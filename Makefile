VERSION ?= dev
LDFLAGS := -X github.com/revazi/tasklight/internal/cli.Version=$(VERSION)

.PHONY: check check-format test test-race vet build run cross-compile macos-helper npm-package package-smoke clean

check:
	./scripts/check.sh

check-format:
	./scripts/check-format.sh

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

build:
	go build -ldflags "$(LDFLAGS)" -o bin/tasklight ./cmd/tasklight

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/tasklight --help

cross-compile:
	./scripts/cross-compile.sh

macos-helper:
	./scripts/build-macos-helper.sh bin darwin-arm64

npm-package:
	npm --prefix npm/tasklight-cli run build:vendor
	npm --prefix npm/tasklight-cli run test:local
	npm --prefix npm/tasklight-cli run pack:check

package-smoke:
	./scripts/package-smoke.sh

clean:
	rm -rf bin coverage.out npm/tasklight-cli/vendor npm/tasklight-cli/*.tgz

BINARY=kubememo
MODULE_DIR=KubeMemo/bin
VERSION?=dev

.PHONY: build test fmt run docs-build docs-serve

build:
	go build -ldflags "-X github.com/KubeDeckio/KubeMemo/internal/cli.version=$(VERSION)" -o $(MODULE_DIR)/$$(go env GOOS)-$$(go env GOARCH)/$(BINARY) ./cmd/kubememo

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

run:
	go run -ldflags "-X github.com/KubeDeckio/KubeMemo/internal/cli.version=$(VERSION)" ./cmd/kubememo

docs-build:
	mkdocs build

docs-serve:
	mkdocs serve

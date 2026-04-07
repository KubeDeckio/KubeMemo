BINARY=kubememo
MODULE_DIR=KubeMemo/bin

.PHONY: build test fmt run docs-build docs-serve

build:
	go build -o $(MODULE_DIR)/$$(go env GOOS)-$$(go env GOARCH)/$(BINARY) ./cmd/kubememo

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

run:
	go run ./cmd/kubememo

docs-build:
	mkdocs build

docs-serve:
	mkdocs serve

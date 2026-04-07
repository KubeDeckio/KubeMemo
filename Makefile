BINARY=kubememo
MODULE_DIR=KubeMemo/bin
VERSION?=dev

.PHONY: build test test-integration fmt run docs-build docs-serve

build:
	go build -ldflags "-X github.com/KubeDeckio/KubeMemo/internal/cli.version=$(VERSION)" -o $(MODULE_DIR)/$$(go env GOOS)-$$(go env GOARCH)/$(BINARY) ./cmd/kubememo

test:
	go test ./...

test-integration:
	KUBEMEMO_INTEGRATION=1 go test ./internal/service -run TestClusterSmokeInstallCreateExportAndCleanup -count=1

fmt:
	gofmt -w ./cmd ./internal

run:
	go run -ldflags "-X github.com/KubeDeckio/KubeMemo/internal/cli.version=$(VERSION)" ./cmd/kubememo

docs-build:
	cat CHANGELOG.md > docs/changelog.md
	mkdocs build

docs-serve:
	cat CHANGELOG.md > docs/changelog.md
	mkdocs serve

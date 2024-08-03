OUT=./build
SRC=.
SRC_CLI=$(SRC)/cli/cli.go
BINARY=gonode

.PHONY: build
build:
	@go build  -o $(OUT)/$(BINARY) $(SRC_CLI)

.PHONY: lint
lint:
	@golangci-lint run ./...
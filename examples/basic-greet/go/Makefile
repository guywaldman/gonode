PKG_NAME=greet

GO=$(shell which go)
SRC=$(PKG_NAME).go
BUILD=./build
LIB_NAME=$(PKG_NAME)

.PHONY: build
build:
	@$(GO) build -buildmode c-archive -o $(BUILD)/$(LIB_NAME).a $(SRC)

.PHONY: generate
generate:
	@$(GO) generate *.go

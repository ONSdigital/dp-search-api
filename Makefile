SHELL=bash
MAIN=dp-search-query

BUILD=build
BUILD_ARCH=$(BUILD)/$(GOOS)-$(GOARCH)
BIN_DIR?=.

export GOOS?=$(shell go env GOOS)
export GOARCH?=$(shell go env GOARCH)

.PHONY: all
all: audit test build

.PHONY: audit
audit:
	nancy go.sum

.PHONY: build
build:
	@mkdir -p $(BUILD_ARCH)/$(BIN_DIR)
	go build -o $(BUILD_ARCH)/$(BIN_DIR)/$(MAIN) cmd/dp-search-query/main.go

.PHONY: debug
debug: build
	HUMAN_LOG=1 go run -race cmd/dp-search-query/main.go

.PHONY: acceptance-publishing
acceptance-publishing: build
	HUMAN_LOG=1 go run cmd/dp-search-query/main.go

.PHONY: acceptance-web
acceptance-web: build
	HUMAN_LOG=1 go run cmd/dp-search-query/main.go

.PHONY: test
test:
	go test -v -cover $(shell go list ./... | grep -v /vendor/)

.PHONY: build debug test

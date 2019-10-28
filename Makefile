SHELL=bash
MAIN=dp-search-query

BUILD=build
BUILD_ARCH=$(BUILD)/$(GOOS)-$(GOARCH)
BIN_DIR?=.

export GOOS?=$(shell go env GOOS)
export GOARCH?=$(shell go env GOARCH)

build:
	@mkdir -p $(BUILD_ARCH)/$(BIN_DIR)
	go build -o $(BUILD_ARCH)/$(BIN_DIR)/$(MAIN) cmd/dp-search-query/main.go
debug: build
	HUMAN_LOG=1 go run -race cmd/dp-search-query/main.go
acceptance-publishing: build
	HUMAN_LOG=1 go run cmd/dp-search-query/main.go
acceptance-web: build
	HUMAN_LOG=1 go run cmd/dp-search-query/main.go
test:
	go test -v -cover $(shell go list ./... | grep -v /vendor/)

.PHONY: build debug test

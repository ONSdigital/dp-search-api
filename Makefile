SHELL=bash
MAIN=dp-search-query

BUILD=build
BUILD_ARCH=$(BUILD)/$(GOOS)-$(GOARCH)
BIN_DIR?=.

export GOOS?=$(shell go env GOOS)
export GOARCH?=$(shell go env GOARCH)

build:
	@mkdir -p $(BUILD_ARCH)/$(BIN_DIR)
	go build -o $(BUILD_ARCH)/$(BIN_DIR)/$(MAIN)
debug: build
	HUMAN_LOG=1 go run -race main.go
acceptance-publishing: build
	ENABLE_PRIVATE_ENDPOINTS=true MONGODB_DATABASE=test HUMAN_LOG=1 go run main.go
acceptance-web: build
	ENABLE_PRIVATE_ENDPOINTS=false MONGODB_DATABASE=test HUMAN_LOG=1 go run main.go
test:
	go test -v -cover $(shell go list ./... | grep -v /vendor/)

.PHONY: build debug test

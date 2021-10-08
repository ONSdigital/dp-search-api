SHELL=bash
MAIN=dp-search-api

BUILD=build
BUILD_ARCH=$(BUILD)/$(GOOS)-$(GOARCH)
BIN_DIR?=.

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)
LDFLAGS=-ldflags "-w -s -X 'main.Version=${VERSION}' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)'"

export GOOS?=$(shell go env GOOS)
export GOARCH?=$(shell go env GOARCH)

.PHONY: all
all: audit test build

.PHONY: audit
audit:
	go list -m all | nancy sleuth

.PHONY: build
build:
	@mkdir -p $(BUILD_ARCH)/$(BIN_DIR)
	go build $(LDFLAGS) -o $(BUILD_ARCH)/$(BIN_DIR)/$(MAIN) main.go

.PHONY: debug
debug: build
	HUMAN_LOG=1 go run $(LDFLAGS) -race main.go

.PHONY: test
test:
	go test -cover -race ./...

.PHONY: test-component
test-component:
	go test -cover -race -coverprofile="coverage.txt" -coverpkg=github.com/ONSdigital/$(MAIN)/... -component

.PHONY: build debug test

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

.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.45.2
	golangci-lint run ./...

.PHONY: fmt
fmt:
	go fmt ./...
	
.PHONY: local
local:
	export ELASTIC_SEARCH_URL=https://localhost:9200; \
	export AWS_TLS_INSECURE_SKIP_VERIFY=true; \
	export AWS_PROFILE=development; \
	export AWS_FILENAME=$(HOME)/.aws/credentials; \
	HUMAN_LOG=1 go run $(LDFLAGS) -race main.go

.PHONY: build
build:
	@mkdir -p $(BUILD_ARCH)/$(BIN_DIR)
	go build $(LDFLAGS) -o $(BUILD_ARCH)/$(BIN_DIR)/$(MAIN) main.go

.PHONY: debug
debug: 
	HUMAN_LOG=1 go run $(LDFLAGS) -race main.go

.PHONY: test
test:
	go test -cover -race ./...

.PHONY: test-component
test-component:
	go test -cover -race -coverprofile="coverage.txt" -coverpkg=github.com/ONSdigital/$(MAIN)/... -component

.PHONY: build debug test

.PHONY: build-reindex
build-reindex:
	@mkdir -p $(BUILD)
	GOOS=linux GOARCH=amd64 go build -tags=aws -ldflags "-w -s" -o $(BUILD)/reindex cmd/reindex/main.go cmd/reindex/aws.go

.PHONY: reindex
reindex:
	HUMAN_LOG=1 go run -ldflags "-w -s" cmd/reindex/main.go cmd/reindex/local.go

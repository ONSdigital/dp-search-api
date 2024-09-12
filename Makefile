SHELL=bash
MAIN=dp-search-api

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

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
all: delimiter-AUDIT audit delimiter-LINTERS lint delimiter-UNIT-TESTS test delimiter-COMPONENT-TESTS test-component delimiter-FINISH ## Runs multiple targets, audit, lint, test and test-component

.PHONY: audit
audit: ## Runs checks for security vulnerabilities on dependencies (including transient ones)
	go list -m all | nancy sleuth

.PHONY: build
build: ## Builds binary of application code and stores in bin directory as dp-search-reindex-api
	@mkdir -p $(BUILD_ARCH)/$(BIN_DIR)
	go build $(LDFLAGS) -o $(BUILD_ARCH)/$(BIN_DIR)/$(MAIN) main.go

.PHONY: convey
convey: ## Runs unit test suite and outputs results on http://127.0.0.1:8080/
	goconvey ./...

.PHONY: debug
debug: ## Used to run code locally
	HUMAN_LOG=1 go run $(LDFLAGS) -race main.go

.PHONY: debug-run
debug-run: ## Used to build and run code locally in debug mode
	HUMAN_LOG=1 DEBUG=1 go run -tags 'debug' -race $(LDFLAGS) main.go

.PHONY: debug-watch
debug-watch: 
	reflex -d none -c ./reflex

.PHONY: delimiter-%
delimiter-%:
	@echo '===================${GREEN} $* ${RESET}==================='

.PHONY: fmt
fmt: ## Run Go formatting on code
	go fmt ./...

.PHONY: lint
lint: ## Used in ci to run linters against Go code
	golangci-lint run ./...

.PHONY: lint-local
lint-local: ## Use locally to run linters against Go code
	golangci-lint run ./...
	
.PHONY: local
local: ## Exports various configurations and then runs the application locally
	export ELASTIC_SEARCH_URL=https://localhost:9200; \
	export AWS_TLS_INSECURE_SKIP_VERIFY=true; \
	export AWS_PROFILE=development; \
	export AWS_FILENAME=$(HOME)/.aws/credentials; \
	HUMAN_LOG=1 go run $(LDFLAGS) -race main.go

.PHONY: test
test: ## Runs unit tests including checks for race conditions and returns coverage
	go test -count=1 -race -cover ./...

.PHONY: test-component
test-component: ## Runs component test suite
	go test -cover -race -coverprofile="coverage.txt" -coverpkg=github.com/ONSdigital/$(MAIN)/... -component

.PHONY: help
help: ## Show help page for list of make targets
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

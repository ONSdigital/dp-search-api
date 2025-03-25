#!/bin/bash -eux

go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5

pushd dp-search-api
  make lint
popd

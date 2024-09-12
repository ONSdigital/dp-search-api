#!/bin/bash -eux

go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0

pushd dp-search-api
  make lint
popd

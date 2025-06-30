#!/bin/bash -eux

go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
npm install -g @redocly/cli

pushd dp-search-api
  make lint
popd

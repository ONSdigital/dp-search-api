#!/bin/bash -eux

cwd=$(pwd)

export GOPATH=$cwd/go

pushd dp-search-api
  make build && mv build/$(go env GOOS)-$(go env GOARCH)/* $cwd/build
  cp -r templates $cwd/build
  cp Dockerfile.concourse $cwd/build
popd

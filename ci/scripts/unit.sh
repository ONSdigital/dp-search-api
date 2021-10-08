#!/bin/bash -eux

export GOPATH=$(pwd)/go

pushd dp-search-api
  make test
popd

#!/bin/bash -eux

export GOPATH=$(pwd)/go

pushd dp-search-query
  make test
popd

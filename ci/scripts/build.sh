#!/bin/bash -eux

cwd=$(pwd)

export GOPATH=$cwd/go

pushd $GOPATH/src/github.com/ONSdigital/dp-search-query
  make build && mv build/$(go env GOOS)-$(go env GOARCH)/* $cwd/build
  cp -r templates $cwd/build
  cp Dockerfile.concourse $cwd/build
popd

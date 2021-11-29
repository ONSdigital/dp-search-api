#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-search-api
  make test-component
popd

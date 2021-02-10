#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-search-api
  make audit
popd 
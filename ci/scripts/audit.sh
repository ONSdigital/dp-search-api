#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-search-query
  make audit
popd 
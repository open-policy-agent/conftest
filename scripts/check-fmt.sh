#!/usr/bin/env bash

CONFTEST_DIR=$(
    dir=$(dirname "${BASH_SOURCE}")/..
    cd $dir
    pwd
)

fmt=$(gofmt -l $CONFTEST_DIR)
if [ -z $fmt ]; then
    exit 0
else
    echo "$fmt"
    exit 1
fi
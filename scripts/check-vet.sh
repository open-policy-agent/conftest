#!/usr/bin/env bash

dir=$(dirname "${BASH_SOURCE}")/..
cd $dir

go vet ./...
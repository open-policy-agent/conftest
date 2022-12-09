#!/bin/bash

# This script validates that bundles can be pushed and pulled
# from a local registry using the registry image from DockerHub
# https://hub.docker.com/_/registry

trap 'cleanup' ERR SIGTERM SIGINT

CONFTEST="./conftest"
CONTAINER_NAME="conftest-push-pull-e2e"

function cleanup() {
    docker rm $CONTAINER_NAME -f >/dev/null 2>&1
    rm -rf tmp
}

# Run the cleanup at the start of the test to ensure the previous
# test run has been successfully torn down.
cleanup

docker run -p 5000:5000 --name $CONTAINER_NAME -d registry
if [ $? != 0 ]; then
    echo "ERROR RUNNING TEST CONTAINER. IS DOCKER INSTALLED?"
    exit 1
fi

# Give the registry container some time to spin up and initialize.
sleep 5

$CONFTEST push localhost:5000/testpush -p examples/data
if [ $? != 0 ]; then
    echo "ERROR PUSHING BUNDLE"
    exit 1
fi

$CONFTEST pull localhost:5000/testpush -p testpush
if [ $? != 0 ]; then
    echo "ERROR PULLING BUNDLE"
    exit 1
fi

$CONFTEST verify -p testpush/examples/data/policy -d testpush/examples/data/exclusions testpush/examples/data/service.yaml
if [ $? != 0 ]; then
    echo "POLICIES WERE NOT SUCCESSFULLY VERIFIED"
    exit 1
fi

$CONFTEST push localhost:5000/testdatadirectory -p examples/data/policy -d examples/data/exclusions
if [ $? != 0 ]; then
    echo "ERROR PUSHING BUNDLE"
    exit 1
fi

$CONFTEST pull localhost:5000/testdatadirectory -p testdatadirectory
if [ $? != 0 ]; then
    echo "ERROR PULLING BUNDLE"
    exit 1
fi

$CONFTEST verify -p testdatadirectory/examples/data/policy -d testdatadirectory/examples/data/exclusions testdatadirectory/examples/data/service.yaml
if [ $? != 0 ]; then
    echo "POLICIES WERE NOT SUCCESSFULLY VERIFIED"
    exit 1
fi

$CONFTEST push localhost:5000/testdataonly -p '' -d examples/data/exclusions
if [ $? != 0 ]; then
    echo "ERROR PUSHING BUNDLE"
    exit 1
fi

$CONFTEST pull localhost:5000/testdataonly -p testdataonly
if [ $? != 0 ]; then
    echo "ERROR PULLING BUNDLE"
    exit 1
fi

$CONFTEST verify -p '' -d testdataonly/examples/data/exclusions
if [ $? != 0 ]; then
    echo "ERROR LOADING DATA BUNDLES"
    exit 1
fi

$CONFTEST push localhost:5000/test-annotations -p tests/annotations
if [ $? != 0 ]; then
    echo "ERROR PUSHING ANNOTATIONS BUNDLE"
    exit 1
fi

$CONFTEST pull localhost:5000/test-annotations -p tmp
if [ $? != 0 ]; then
    echo "ERROR PULLING ANNOTATIONS BUNDLE"
    exit 1
fi

$CONFTEST verify -p tmp/tests/annotations/policy -d tmp/tests/annotations/exclusions tmp/tests/annotations/service.yaml
if [ $? != 0 ]; then
    echo "POLICIES WITH ANNOTATIONS WERE NOT SUCCESSFULLY VERIFIED"
    exit 1
fi

cleanup
exit 0

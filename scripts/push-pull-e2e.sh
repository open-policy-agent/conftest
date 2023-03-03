#!/bin/env bash

set -o errexit
set -o pipefail
set -o nounset
set -o errtrace

# This script validates that bundles can be pushed and pulled
# from a local registry using the registry image from DockerHub
# https://hub.docker.com/_/registry

trap 'cleanup' EXIT

WORKDIR="$(mktemp -d conftest-push-pull-e2e.XXXXXXXXXX)"
CONFTEST="./conftest"
CONTAINER_NAME="conftest-push-pull-e2e"

function cleanup() {
    docker rm $CONTAINER_NAME -f >/dev/null 2>&1
    rm -rf "${WORKDIR}"
}

# Run the cleanup at the start of the test to ensure the previous
# test run has been successfully torn down.
cleanup

if ! docker run -p 5000:5000 --name $CONTAINER_NAME -d registry; then
    echo "ERROR RUNNING TEST CONTAINER. IS DOCKER INSTALLED?"
    exit 1
fi

# Wait until registry is listening
until [[ $(docker logs ${CONTAINER_NAME} 2>&1) == *"listening on"* ]]; do
    sleep .1
done

if ! $CONFTEST push localhost:5000/testpush -p examples/data; then
    echo "ERROR PUSHING BUNDLE"
    exit 1
fi

if ! $CONFTEST pull localhost:5000/testpush -p "${WORKDIR}"/testpush; then
    echo "ERROR PULLING BUNDLE"
    exit 1
fi

if ! $CONFTEST verify -p "${WORKDIR}"/testpush/examples/data/policy -d "${WORKDIR}"/testpush/examples/data/exclusions "${WORKDIR}"/testpush/examples/data/service.yaml; then
    echo "POLICIES WERE NOT SUCCESSFULLY VERIFIED"
    exit 1
fi

if ! $CONFTEST push localhost:5000/testdatadirectory -p examples/data/policy -d examples/data/exclusions; then
    echo "ERROR PUSHING BUNDLE"
    exit 1
fi

if ! $CONFTEST pull localhost:5000/testdatadirectory -p "${WORKDIR}"/testdatadirectory; then
    echo "ERROR PULLING BUNDLE"
    exit 1
fi

if ! $CONFTEST verify -p "${WORKDIR}"/testdatadirectory/examples/data/policy -d "${WORKDIR}"/testdatadirectory/examples/data/exclusions "${WORKDIR}"/testdatadirectory/examples/data/service.yaml; then
    echo "POLICIES WERE NOT SUCCESSFULLY VERIFIED"
    exit 1
fi

if ! $CONFTEST push localhost:5000/testdataonly -p '' -d examples/data/exclusions; then
    echo "ERROR PUSHING BUNDLE"
    exit 1
fi

if ! $CONFTEST pull localhost:5000/testdataonly -p "${WORKDIR}"/testdataonly; then
    echo "ERROR PULLING BUNDLE"
    exit 1
fi

if ! $CONFTEST verify -p '' -d "${WORKDIR}"/testdataonly/examples/data/exclusions; then
    echo "ERROR LOADING DATA BUNDLES"
    exit 1
fi

if ! $CONFTEST push localhost:5000/test-annotations -p tests/annotations; then
    echo "ERROR PUSHING ANNOTATIONS BUNDLE"
    exit 1
fi

if ! $CONFTEST pull localhost:5000/test-annotations -p "${WORKDIR}"; then
    echo "ERROR PULLING ANNOTATIONS BUNDLE"
    exit 1
fi

if ! $CONFTEST verify -p "${WORKDIR}"/tests/annotations/policy -d "${WORKDIR}"/tests/annotations/exclusions "${WORKDIR}"/tests/annotations/service.yaml; then
    echo "POLICIES WITH ANNOTATIONS WERE NOT SUCCESSFULLY VERIFIED"
    exit 1
fi

# stop the unauthenticated docker registry
docker rm $CONTAINER_NAME -f >/dev/null 2>&1

# create key and certificate, TLS is requirement for basic auth
mkdir -p "${WORKDIR}"/certs
openssl req \
  -newkey rsa:4096 \
  -nodes \
  -sha256 \
  -batch \
  -subj "/CN=registry.127.0.0.1.nip.io" \
  -addext "subjectAltName = DNS:registry.127.0.0.1.nip.io" \
  -x509 \
  -days 365 \
  -keyout "${WORKDIR}"/certs/domain.key \
  -out "${WORKDIR}"/certs/domain.crt >/dev/null 2>&1

# create htpasswd file with the test user
mkdir -p "${WORKDIR}"/auth
docker run \
  --entrypoint htpasswd \
  --rm \
  httpd:2 \
  -Bbn \
  test supersecret > "${WORKDIR}"/auth/htpasswd

if ! docker run \
  -d \
  -p 5000:5000 \
  --name $CONTAINER_NAME \
  -v "${PWD}/${WORKDIR}"/auth:/auth:Z \
  -e "REGISTRY_AUTH=htpasswd" \
  -e "REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm" \
  -e REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd \
  -v "${PWD}/${WORKDIR}"/certs:/certs:Z \
  -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/domain.crt \
  -e REGISTRY_HTTP_TLS_KEY=/certs/domain.key \
  registry; then
    echo "ERROR STARTING TEST REGISTRY CONTAINER"
    exit 1
fi

# Wait until registry is listening
until [[ $(docker logs ${CONTAINER_NAME} 2>&1) == *"listening on"* ]]; do
    sleep .1
done

export SSL_CERT_FILE="${PWD}/${WORKDIR}"/certs/domain.crt

if ! echo supersecret | docker login --username=test --password-stdin registry.127.0.0.1.nip.io:5000 >/dev/null 2>&1; then
    echo "LOGGING TO TEST REGISTRY"
    exit 1
fi

if ! $CONFTEST push --tls oci://registry.127.0.0.1.nip.io:5000/testdataonly -p examples/data/policy -d examples/data/exclusions; then
    echo "PUSHING TO REGISTRY VIA TLS AND WITH AUTHENTICATION FAILED"
    exit 1
fi

if ! $CONFTEST pull --tls oci://registry.127.0.0.1.nip.io:5000/testdataonly -p "${WORKDIR}"/tlstest; then
    echo "PULLING FROM REGISTRY VIA TLS AND WITH AUTHENTICATION FAILED"
    sleep infinity
    exit 1
fi

if ! $CONFTEST verify -p "${WORKDIR}"/tlstest -d "${WORKDIR}"/tlstest/examples/data/exclusions examples/data/service.yaml; then
    echo "POLICIES PULLED FROM TLS REGISTRY WITH AUTHENTICATION NOT SUCCESSFULLY VERIFIED"
    exit 1
fi

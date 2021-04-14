ROOT_DIR := ../..

OS := $(if $(GOOS),$(GOOS),$(shell go env GOOS))

BIN_EXTENSION :=
ifeq ($(OS), windows)
  BIN_EXTENSION := .exe
endif

BIN := conftest$(BIN_EXTENSION)

IMAGE := openpolicyagent/conftest

TAG := $(shell git describe --abbrev=0 --tags)
GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_TAG := $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
DATE := $(shell date)

VERSION = unreleased
ifneq ($(GIT_TAG),)
	VERSION = $(GIT_TAG)
endif

DOCKER := DOCKER_BUILDKIT=1 docker

## All of the directories that contain tests to be executed
## e.g. echo $(TEST_DIRS) prints tests/foo tests/bar
TEST_DIRS := $(patsubst tests/%/, tests/%, $(dir $(wildcard tests/**/.)))

#
##@ Development
#

.PHONY: build
build: ## Builds Conftest.
	@go build

.PHONY: test
test: ## Tests Conftest.
	@go test -v ./...

.PHONY: test-examples
test-examples: build ## Runs the tests for the examples.
	@bats acceptance.bats

.PHONY: test-acceptance
test-acceptance: build ## Runs the tests in the test folder.
	@for testdir in $(TEST_DIRS) ; do \
		cd $(CURDIR)/$$testdir && CONFTEST=$(ROOT_DIR)/$(BIN) bats test.bats || exit 1; \
	done

.PHONY: lint
lint: ## Lints Conftest.
	@golint -set_exit_status ./...
	@go vet ./...

.PHONY: all
all: lint build test test-examples test-acceptance ## Runs all linting and tests.

help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

#
##@ Releases
#

.PHONY: image
image: ## Builds a Docker image for Conftest.
	@$(DOCKER) build --build-arg VERSION="$(VERSION)" --build-arg COMMIT="$(GIT_COMMIT)" --build-arg DATE="$(DATE)" . -t $(IMAGE):$(TAG) 
	@$(DOCKER) tag $(IMAGE):$(TAG) $(IMAGE):latest

.PHONY: examples
examples: ## Builds the examples Docker image.
	@$(DOCKER) build . --target examples -t $(IMAGE):examples

.PHONY: push
push: examples image ## Pushes the examples and Conftest image to DockerHub.
	@$(DOCKER) push $(IMAGE):$(TAG)
	@$(DOCKER) push $(IMAGE):latest
	@$(DOCKER) push $(IMAGE):examples

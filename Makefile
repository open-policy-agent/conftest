ROOT_DIR := ../../..

OS := $(if $(GOOS),$(GOOS),$(shell go env GOOS))

BIN_EXTENSION :=
ifeq ($(OS), windows)
  BIN_EXTENSION := .exe
endif

BIN := conftest$(BIN_EXTENSION)

## All of the test directories specific to issues
## e.g. echo $(ISSUE_TEST_DIRS) prints tests/issues/000 tests/issues/001
ISSUE_TEST_DIRS := $(patsubst tests/%/, tests/%, $(dir $(wildcard tests/**/**/.)))

.PHONY: build
build:
	@go build

.PHONY: test
test:
	@go test -v ./...

.PHONY: acceptance
acceptance: build
	@bats acceptance.bats

.PHONY: test-issues
test-issues: build
	@for testdir in $(ISSUE_TEST_DIRS) ; do \
		cd $(CURDIR)/$$testdir && CONFTEST=$(ROOT_DIR)/$(BIN) bats test.bats ; \
	done

.PHONY: all
all: build test acceptance test-issues

## RELEASES
TAG=$(shell git describe --abbrev=0 --tags)
IMAGE=openpolicyagent/conftest
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_TAG=$(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
DATE=$(shell date)

VERSION = unreleased
ifneq ($(GIT_TAG),)
	VERSION = $(GIT_TAG)
endif

.PHONY: image
image:
	@docker build --build-arg VERSION="$(VERSION)" --build-arg COMMIT="$(GIT_COMMIT)" --build-arg DATE="$(DATE)" . -t $(IMAGE):$(TAG) 
	@docker tag $(IMAGE):$(TAG) $(IMAGE):latest

.PHONY: examples
examples:
	@docker build . --target examples -t $(IMAGE):examples

.PHONY: push
push: examples image
	@docker push $(IMAGE):$(TAG)
	@docker push $(IMAGE):latest
	@docker push $(IMAGE):examples

.PHONY: check-vet
check-vet: 
	@go vet ./...

.PHONY: check-lint
check-lint:
	@golint -set_exit_status ./...

.PHONY: check
check: check-vet check-lint

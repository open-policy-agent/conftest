## DEVELOPMENT
.PHONY: build
build: 
	@go build

.PHONY: test
test: 
	@go test -v ./...

.PHONY: acceptance
acceptance: 
	@bats acceptance.bats

.PHONY: all
all: build test acceptance

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

check-vet:
	@go vet ./...

check-lint:
	@golint -set_exit_status ./...

check: check-vet check-lint

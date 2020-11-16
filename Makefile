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

.PHONY: image
image:
	@docker build . -t $(IMAGE):$(TAG)
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

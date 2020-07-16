
TAG=$(shell git describe --abbrev=0 --tags)

NAME=conftest
IMAGE=openpolicyagent/$(NAME)
ALT_IMAGE=instrumenta/$(NAME)

COMMAND=docker
BUILD=DOCKER_BUILDKIT=1 $(COMMAND) build --pull
PUSH=$(COMMAND) push

all: push

examples:
	$(BUILD) --target examples -t $(IMAGE):examples .
	$(COMMAND) tag $(IMAGE):examples $(ALT_IMAGE):examples

acceptance:
	$(BUILD) --target acceptance .

conftest:
	$(BUILD) -t $(IMAGE):$(TAG) .
	$(COMMAND) tag $(IMAGE):$(TAG) $(IMAGE):latest
	$(COMMAND) tag $(IMAGE):$(TAG) $(ALT_IMAGE):latest
	$(COMMAND) tag $(IMAGE):$(TAG) $(ALT_IMAGE):$(TAG)

test: conftest
	$(BUILD) --target test .

push: examples conftest
	$(PUSH) $(IMAGE):$(TAG)
	$(PUSH) $(IMAGE):latest
	$(PUSH) $(ALT_IMAGE):$(TAG)
	$(PUSH) $(ALT_IMAGE):latest
	$(PUSH) $(IMAGE):examples
	$(PUSH) $(ALT_IMAGE):examples

.PHONY: examples acceptance conftest push all

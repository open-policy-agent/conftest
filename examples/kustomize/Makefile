all: test

test:
	kustomize build | conftest -

show:
	kustomize build

.PHONY: all, test, show

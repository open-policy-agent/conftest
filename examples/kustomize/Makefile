all: test

test:
	kustomize build | conftest test -

show:
	kustomize build

.PHONY: all, test, show

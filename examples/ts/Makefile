FILE := pod.ts

all: test

node_modules:
	npm install --dev

test: node_modules
	npx ts-node $(FILE) | conftest test -

show:
	@npx ts-node $(FILE)

.PHONY: all, test, show

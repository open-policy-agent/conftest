# Development

This document highlights the required tools and workflows to develop for Conftest.

## Tools

### Go

Conftest is written in the [Go](https://golang.org) programming language, and can be installed from their [installation page](https://golang.org/doc/install).

If you are not familiar with Go we recommend you read through the [How to Write Go Code](https://golang.org/doc/code.html) article to familiarize yourself with the standard Go development environment.

### Make

[Make](https://www.gnu.org/software/make/) is used for local development and assists with running the builds and tests.

Windows users can download Make from [here](http://gnuwin32.sourceforge.net/packages/make.htm) if not already installed.

### Bats

[Bats](https://github.com/sstephenson/bats) is used for running the [acceptance tests](acceptance.bats).

There are a few ways to install Bats:

- Brew: `brew install bats-core`
- npm: `npm install -g bats`

### GolangCI-lint

[golangci-lint](https://golangci-lint.run/) is a Go linters aggregator and is used for running lint tasks.

## Building and Testing

All build and testing workflows have `make` commands.

- Build: `make build`

- Test: `make test`

- Acceptance: `make test-acceptance`

- Run everything! `make all`

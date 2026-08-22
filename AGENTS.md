# AGENTS.md

This file gives AI coding assistants the minimum context they need to work
productively in this repository. Human contributors should start with
[CONTRIBUTING.md](CONTRIBUTING.md) and [DEVELOPMENT.md](DEVELOPMENT.md), which
remain the source of truth.

## Project at a glance

Conftest is a CLI for writing tests against structured configuration data
(Kubernetes manifests, Terraform, Dockerfiles, etc.) using
[Rego](https://www.openpolicyagent.org/docs/policy-language) from the Open
Policy Agent project. User documentation lives at
[conftest.dev](https://www.conftest.dev/).

## Repository layout

- `cmd/` — Cobra subcommands (`test`, `verify`, `parse`, `pull`, `push`, ...).
- `parser/` — input parsers per format (yaml, json, hcl, toml, dockerfile, ...).
- `output/` — output formatters (`stdout`, `json`, `tap`, `table`, `junit`,
  `sarif`, ...).
- `policy/` — policy loading, compilation, and evaluation.
- `runner/` — orchestrates parsing, policy evaluation, and output.
- `plugin/` — plugin discovery and execution.
- `examples/` — end-to-end examples per input format and feature.
- `tests/` — bats-based acceptance fixtures used by `acceptance.bats`.
- `acceptance.bats` — top-level acceptance suite.
- `internal/` — private helpers; do not import from outside the module.

## Building and testing

- Unit tests: `make test` (equivalent to `go test ./...`).
- Acceptance tests: `make test-acceptance` (runs `bats acceptance.bats`; needs
  [bats-core](https://github.com/bats-core/bats-core)).
- Lint: `make lint` (runs `golangci-lint`).
- Build: `make build`.
- Run everything: `make all`.

When adding behaviour to a parser, output, or command, prefer adding both a Go
unit test and a small example under `examples/` plus an acceptance case in
`tests/` if the feature is user-visible.

## Pull request expectations

- Use [conventional commit](https://www.conventionalcommits.org/) prefixes
  (`feat:`, `fix:`, `docs:`, `chore:`, ...) — enforced by repo conventions.
- Sign off every commit (`git commit -s`); the project requires DCO.
- Squash related commits before merge.
- Most behaviour changes should ship with tests.
- Reference the issue with `Fixes #<n>` in the commit body when applicable.

## Things to avoid

- Do not commit binaries, caches, or local Rego output.
- Do not import from `internal/` outside the module.
- Do not add new top-level CLI flags without a matching `examples/` and
  acceptance case.
- Do not change the Rego query semantics without a deprecation plan; users
  pin Conftest to specific versions in CI.

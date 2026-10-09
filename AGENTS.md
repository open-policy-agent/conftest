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

- `internal/commands/`: Cobra subcommands (`test`, `verify`, `parse`, `pull`,
  `push`, ...). There is no top-level `cmd/` directory.
- `parser/`: input parsers per format (yaml, json, hcl, toml, dockerfile, ...).
- `output/`: output formatters (`stdout`, `json`, `tap`, `table`, `junit`,
  `sarif`, ...).
- `policy/`: policy loading, compilation, and evaluation.
- `builtins/`: custom Rego builtins exposed to policies.
- `downloader/`: fetching policies and data from remote sources.
- `document/`: generates policy documentation from Rego metadata.
- `runner/`: orchestrates parsing, policy evaluation, and output.
- `plugin/`: plugin discovery and execution.
- `docs/`: the conftest.dev site sources; formatted with
  `make format-docs` (`mdformat --wrap 80 docs`).
- `examples/`: end-to-end examples per input format and feature, exercised by
  `acceptance.bats`.
- `tests/`: one directory per acceptance case, each containing a `test.bats`.
- `acceptance.bats`: top-level suite covering `examples/`.
- `internal/`: private packages; do not import from outside the module.

## Building and testing

- Unit tests: `make test` (`go test -v ./...`).
- Example tests: `make test-examples` builds, then runs `bats acceptance.bats`
  against `examples/`.
- Acceptance tests: `make test-acceptance` builds, then iterates every
  directory under `tests/` running its `test.bats` with `CONFTEST` pointing at
  the freshly built binary.
- Lint: `make lint` (`golangci-lint run --fix`).
- Build: `make build`.
- Run everything: `make all` (lint, build, test, test-examples, test-acceptance).

CI runs more than `make lint`. The PR workflow also runs
[regal](https://github.com/StyraInc/regal) over `examples/` (warnings are
non-blocking, the `bugs` category is blocking) and runs `ratchet` to validate
that workflow action refs are pinned.

Both bats suites need [bats-core](https://github.com/bats-core/bats-core).

## Development environment

The repository ships a Nix flake. `nix develop` gives you a shell with the
tools the Makefile and CI expect, including `bats`, `golangci-lint`, `ratchet`,
`regal`, `mdformat`, `goreleaser` and the Go toolchain pinned from `go.mod`.
`make ratchet-update`, `make format-docs` and the regal lint step depend on
tools that are only available there, so prefer the dev shell over installing
them piecemeal.

When adding behaviour to a parser, output, or command, prefer adding both a Go
unit test and a small example under `examples/` plus an acceptance case in
`tests/` if the feature is user-visible.

## Pull request expectations

- Use [conventional commit](https://www.conventionalcommits.org/) prefixes
  (`feat:`, `fix:`, `docs:`, `chore:`, ...). CI validates the commit header, and
  the scope must be a single word: `feat(test):` is accepted, `feat(test,verify):`
  is not.
- Sign off every commit (`git commit -s`); the project requires DCO.
- Squash related commits before merge.
- Most behaviour changes should ship with tests.
- Reference the issue with `Fixes #<n>` in the commit body when applicable.
- Workflow action refs are SHA-pinned with trailing `# ratchet:` comments and CI
  fails on unpinned refs. After editing anything in `.github/workflows/`, run
  `make ratchet-update`.

## Things to avoid

- Do not commit binaries, caches, or local Rego output.
- Do not import from `internal/` outside the module.
- Do not add new top-level CLI flags without a matching `examples/` and
  acceptance case.
- Do not change the Rego query semantics without a deprecation plan; users
  pin Conftest to specific versions in CI.

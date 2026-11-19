# AGENTS.md

This file guides AI coding agents (e.g. Cursor, Copilot, Claude) on how to
contribute to the Conftest project.

## Project Overview

Conftest is a utility for testing structured configuration files (Kubernetes
manifests, Terraform plans, Dockerfile, etc.) using policies written in
[Rego](https://www.openpolicyagent.org/docs/latest/policy-language/) (the
Open Policy Agent policy language).

- Website: <https://www.conftest.dev>
- OPA docs: <https://www.openpolicyagent.org/docs/latest/>

## Repository Layout

```
conftest/
├── main.go              # CLI entry point (uses cobra)
├── runner/              # Core test runner logic
├── parser/              # File parsers (YAML, JSON, TOML, HCL, ...)
├── output/              # Output formatters (stdout, JSON, TAP, ...)
├── policy/              # Rego policy loading and evaluation
├── downloader/          # Bundle/policy download from remote sources
├── plugin/              # Plugin support
├── internal/            # Internal helpers
├── examples/            # End-to-end usage examples (great for reference!)
├── tests/               # Integration tests
├── acceptance.bats      # Bats acceptance tests
├── docs/                # Documentation source (MkDocs)
└── Makefile             # All common workflows
```

## Development Environment

### Prerequisites

- **Go** ≥ 1.22 – <https://go.dev/doc/install>
- **Make** – standard build driver
- **Bats** (for acceptance tests) – `brew install bats-core` or `npm i -g bats`
- **golangci-lint** (for linting) – <https://golangci-lint.run/usage/install/>

Optional but useful:
- **Nix** – `nix develop` brings up a reproducible dev environment (see `flake.nix`)

### Build & Test Workflow

```bash
# Build the binary
make build

# Run unit tests
go test ./...
# or
make test

# Run acceptance tests (requires the built binary on $PATH)
make test-acceptance
# or directly
bats acceptance.bats

# Lint
make lint

# Run all checks (build + test + lint)
make ci
```

### Local smoke test

```bash
# Build and run against the bundled examples
make build
./conftest test --help

# Try one of the examples
./conftest test examples/kubernetes/deployment.yaml -p examples/kubernetes/policy
```

## Writing Tests

### Unit tests

Unit tests live alongside their source files (`_test.go`). Follow existing
patterns. Run with `go test ./...`.

### Acceptance tests

`acceptance.bats` uses the [Bats](https://bats-core.readthedocs.io/) framework.
Each test calls the compiled `conftest` binary. Add new test cases at the end
of the file, following the existing `@test` style:

```bash
@test "my new feature works" {
  cd "$TESTS_DIR"
  run conftest test --my-flag examples/kubernetes/deployment.yaml -p examples/kubernetes/policy
  [ "$status" -eq 0 ]
  [[ "$output" == *"expected text"* ]]
}
```

### Examples

The `examples/` directory contains self-contained scenarios for each
supported file type. When adding a new parser or feature, add a
corresponding example under `examples/<feature>/`.

## Code Conventions

- Standard Go conventions apply (`gofmt`, `go vet`)
- Errors should be wrapped with `fmt.Errorf("context: %w", err)`
- CLI flags are added to the cobra command in the relevant `cmd/*.go` file
- New parsers implement the `parser.Parser` interface in `parser/`
- New output formatters implement the `output.Outputter` interface in `output/`

## Submitting a PR

1. Fork the repo and create a feature branch.
2. Make your changes and add tests.
3. Run `go test ./...` and `bats acceptance.bats` — both must pass.
4. Open a PR against the `master` branch with a clear description.

See [CONTRIBUTING.md](CONTRIBUTING.md) for full guidelines.

# AGENTS.md

Guidance for AI coding agents contributing to Conftest.

## Project

Conftest helps you write tests against structured configuration data using the
Rego language from Open Policy Agent. See https://www.conftest.dev/ for full
documentation.

## Building and testing

- Run the Go unit tests: `go test ./...`
- Run the acceptance tests: `bats acceptance.bats`

## Development environment

If you use Nix, the dev environment can be brought up with `nix develop`.

## Adding tests and examples

Look at the existing `tests/` and `examples/` directories for patterns to follow
when generating similar tests or scenarios.

See `DEVELOPMENT.md` and `CONTRIBUTING.md` for more details.

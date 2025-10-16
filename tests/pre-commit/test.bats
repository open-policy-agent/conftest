#!/usr/bin/env bats

DIR="$( cd "$( dirname "${BATS_TEST_FILENAME}" )" >/dev/null 2>&1 && pwd )"
PROJECT_ROOT="$( cd "$DIR/../.." >/dev/null 2>&1 && pwd )"

# Git configuration for temporary repo
GIT_AUTHOR_NAME="Conftest Test User"
GIT_AUTHOR_EMAIL="conftest@example.tld"

setup_file() {
    # Create a temporary directory for testing
    export TEST_REPO=$(mktemp -d)
    cd "$TEST_REPO"

    # Initialize a new Git repository
    git init

    # Configure Git to use environment variables and disable signing
    git config commit.gpgsign false
    git config tag.gpgsign false
    git config user.name "$GIT_AUTHOR_NAME"
    git config user.email "$GIT_AUTHOR_EMAIL"

    # Copy necessary files from the main repo
    mkdir -p examples
    cp -r "$PROJECT_ROOT/examples/kubernetes" examples/

    # Create pre-commit config
    cat > .pre-commit-config.yaml << EOF
repos:
- repo: ${PROJECT_ROOT}
  rev: HEAD
  hooks:
    - id: conftest-test
      args:
        - --policy
        - examples/kubernetes/policy
    - id: conftest-verify
      args:
        - --policy
        - examples/kubernetes/policy
    # This hook is intended to change/fmt this file
    - id: conftest-fmt
      files: examples/kubernetes/deployment.yaml
    - id: conftest-pull
      args:
        - --policy
        - ./pulled-policies
        - git::https://github.com/open-policy-agent/conftest//examples/kubernetes/policy
EOF

    # Add and commit files
    git add .
    git commit -m "Initial commit"

    # Install pre-commit hooks in the temporary repo
    run pre-commit try-repo "$PROJECT_ROOT"
    run pre-commit install --hook-type pre-commit
    [ "$status" -eq 0 ]
}

teardown_file() {
    # Clean up the temporary repository
    rm -rf "$TEST_REPO"
}

@test "pre-commit: test hook validates as expected" {
    cd "$TEST_REPO"
    run pre-commit run conftest-test --files examples/kubernetes/deployment.yaml
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Containers must not run as root" ]]
}

@test "pre-commit: verify hook runs policy tests" {
    cd "$TEST_REPO"
    run pre-commit run conftest-verify
    [ "$status" -eq 0 ]
}

@test "pre-commit: verify fmt hook changes a policy file" {
    cd "$TEST_REPO"
    run pre-commit run conftest-fmt --files examples/kubernetes/deployment.yaml
    [ "$status" -ne 0 ]
}

@test "pre-commit: pull hook downloads policies successfully" {
  cd "$TEST_REPO"
  run pre-commit run conftest-pull
  [ "$status" -eq 0 ]
}



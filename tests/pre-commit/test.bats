#!/usr/bin/env bats

DIR="$( cd "$( dirname "${BATS_TEST_FILENAME}" )" >/dev/null 2>&1 && pwd )"

setup_file() {
    cd "$DIR/../.."
    # Verify pre-commit is installed and in PATH
    which pre-commit
    pre-commit --version
    cat > .pre-commit-config.yaml << 'EOF'
repos:
- repo: .
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
EOF
    # Update ref to latest commit
    run pre-commit autoupdate --bleeding-edge
    # Install the pre-commit hooks
    run pre-commit try-repo .
    run pre-commit install --install-hooks
    [ "$status" -eq 0 ]
}

teardown_file() {
    cd "$DIR/../.."
    # Remove the pre-commit config file created during setup
    rm -f .pre-commit-config.yaml

    # Uninstall the pre-commit hooks
    run pre-commit uninstall
    [ "$status" -eq 0 ]
}

@test "pre-commit: test hook validates as expected" {
    cd "$DIR/../.."
    run pre-commit run conftest-test --files examples/kubernetes/deployment.yaml
    [ "$status" -eq 1 ]
    [[ "$output" =~ "Containers must not run as root" ]]
}

@test "pre-commit: verify hook runs policy tests" {
    cd "$DIR/../.."
    run pre-commit run conftest-verify
    [ "$status" -eq 0 ]
}

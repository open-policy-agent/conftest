#!/usr/bin/env bats

setup() {
    # Create temporary directories for testing
    export TEMP_DIR=$(mktemp -d)
    export REL_TEMP_DIR="examples/tmp-conftest-test-$$"
    export ABS_POLICY_DIR="${TEMP_DIR}/conftest-policies"
    mkdir -p "${ABS_POLICY_DIR}"
    mkdir -p "${REL_TEMP_DIR}"
}

teardown() {
    # Clean up temporary directories
    rm -rf "${TEMP_DIR}"
    rm -rf "${REL_TEMP_DIR}"
}

@test "Pull command works with relative paths (default behavior)" {
    run $CONFTEST pull --policy $REL_TEMP_DIR https://raw.githubusercontent.com/open-policy-agent/conftest/master/examples/compose/policy/deny.rego
    [ "$status" -eq 0 ]
    [ -d "$REL_TEMP_DIR" ]
    [ -f "$REL_TEMP_DIR/deny.rego" ]
}

@test "Pull command uses absolute paths as relative when --absolute-paths is not set" {
    run $CONFTEST pull --policy "${ABS_POLICY_DIR}" https://raw.githubusercontent.com/open-policy-agent/conftest/master/examples/compose/policy/deny.rego
    [ "$status" -eq 0 ]
    # The policy should be downloaded to ./ABS_POLICY_DIR instead of the absolute path
    [ ! -d "${ABS_POLICY_DIR}/deny.rego" ]
    [ -f "./${ABS_POLICY_DIR}/deny.rego" ]
}

@test "Pull command works with absolute path when --absolute-paths is set" {
    run $CONFTEST pull --absolute-paths --policy "${ABS_POLICY_DIR}" https://raw.githubusercontent.com/open-policy-agent/conftest/master/examples/compose/policy/deny.rego
    [ "$status" -eq 0 ]
    # The policy should be downloaded to the absolute path
    [ ! -f "./${ABS_POLICY_DIR#/}/deny.rego" ]
    [ -f "${ABS_POLICY_DIR}/deny.rego" ]
}
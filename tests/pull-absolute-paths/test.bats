#!/usr/bin/env bats

setup() {
    # Create temporary directories for testing
    export TEMP_DIR=$(mktemp -d)
    export REL_TEMP_DIR="examples/tmp-conftest-test-$$"

    # On Windows (MSYS2/Git Bash), convert to mixed-style path for conftest compatibility
    if command -v cygpath >/dev/null 2>&1; then
        TEMP_DIR=$(cygpath -m "${TEMP_DIR}")
    fi

    export ABS_POLICY_DIR="${TEMP_DIR}/conftest-policies"
    mkdir -p "${ABS_POLICY_DIR}"
    mkdir -p "${REL_TEMP_DIR}"
}

teardown() {
    # Clean up temporary directories
    rm -rf "${TEMP_DIR}"
    rm -rf "${REL_TEMP_DIR}"

    # Clean up any relative path directories created by conftest (test 2)
    # On Windows, the path has drive letter stripped, on Unix it's just the path
    if command -v cygpath >/dev/null 2>&1; then
        # Windows: strip drive letter (e.g., C:/Users/... -> /Users/...)
        local rel_path="${ABS_POLICY_DIR#[A-Za-z]:}"
        rm -rf ".${rel_path}"
    else
        rm -rf ".${ABS_POLICY_DIR}"
    fi
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

    # The policy should be downloaded to a relative path (./path stripped of volume)
    # On Windows: C:/Users/... becomes ./Users/...
    # On Unix: /tmp/... becomes ./tmp/...
    if command -v cygpath >/dev/null 2>&1; then
        local expected_path=".${ABS_POLICY_DIR#[A-Za-z]:}"
    else
        local expected_path=".${ABS_POLICY_DIR}"
    fi
    [ -f "${expected_path}/deny.rego" ]
}

@test "Pull command works with absolute path when --absolute-paths is set" {
    run $CONFTEST pull --absolute-paths --policy "${ABS_POLICY_DIR}" https://raw.githubusercontent.com/open-policy-agent/conftest/master/examples/compose/policy/deny.rego
    [ "$status" -eq 0 ]
    # The policy should be downloaded to the absolute path
    [ -f "${ABS_POLICY_DIR}/deny.rego" ]
}

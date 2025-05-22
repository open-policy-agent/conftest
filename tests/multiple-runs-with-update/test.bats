#!/usr/bin/env bats

DIR="$( cd "$( dirname "${BATS_TEST_FILENAME}" )" >/dev/null 2>&1 && pwd )"
REMOTE_POLICY_FILE="file::${DIR}/remote-policy/policy.rego"

@test "First run of policy fetched with --update flag" {
  run $CONFTEST test -p policy --update ${REMOTE_POLICY_FILE} file.json
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2 tests, 2 passed, 0 warnings, 0 failures" ]]
}

@test "Second run of policy fetched with --update flag" {
  run $CONFTEST test -p policy --update ${REMOTE_POLICY_FILE} file.json
  [ "$status" -eq 0 ]
  [[ "$output" =~ "2 tests, 2 passed, 0 warnings, 0 failures" ]]
}

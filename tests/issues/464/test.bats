#!/usr/bin/env bats

@test "Policy with multiple failures returns a postive number of failures" {
  run $CONFTEST test -p policy.rego file.json
  [ "$status" -eq 1 ]
  [[ "$output" =~ "3 tests, 0 passed, 0 warnings, 3 failures" ]]
}

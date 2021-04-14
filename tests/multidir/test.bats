#!/usr/bin/env bats

@test "Can have multiple policy flags" {
  run $CONFTEST test --policy org --policy team data.json

  [ "$status" -eq 1 ]
  [[ "$output" =~ "2 tests, 0 passed, 0 warnings, 2 failures" ]]
}

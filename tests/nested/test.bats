#!/usr/bin/env bats

@test "Can parse nested files with name overlap (first)" {
  run $CONFTEST test --namespace group1 data.json
  [ "$status" -eq 1 ]
}

@test "Can parse nested files with name overlap (second)" {
  run $CONFTEST test --namespace group2 data.json
  [ "$status" -eq 1 ]
}

@test "Can have multiple namespace flags" {
  run $CONFTEST test --namespace group1 --namespace group2 data.json

  [ "$status" -eq 1 ]
  [[ "$output" =~ "2 tests, 0 passed, 0 warnings, 2 failures" ]]
}

#!/usr/bin/env bats

@test "Parsing error without show-builtin-errors flag returns test failed" {
  run $CONFTEST verify --show-builtin-errors=false

  [ "$status" -eq 1 ]
  echo $output
  [[ "$output" =~ "1 test, 0 passed, 0 warnings, 1 failure, 0 exceptions, 0 skipped" ]]
}

@test "Parsing error with show-builtin-errors flag returns builtin error" {
  run $CONFTEST verify --show-builtin-errors=true

  [ "$status" -eq 1 ]
  echo $output
  [[ "$output" =~ "file_does_not_exist.yaml: no such file or directory" ]]
}

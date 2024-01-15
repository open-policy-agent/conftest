#!/usr/bin/env bats

@test "Test works as expected using contains and if" {
  run $CONFTEST test --policy=policy/valid.rego data.yaml

  [ "$status" -eq 1 ]
  echo $output
  [[ "$output" =~ "1 test, 0 passed, 0 warnings, 1 failure, 0 exceptions" ]]
}

@test "Error is raised when if is used without contains" {
  run $CONFTEST test --policy=policy/invalid.rego data.yaml

  [ "$status" -eq 1 ]
  echo $output
  [[ "$output" =~ "'if' keyword without 'contains' keyword" ]]
}

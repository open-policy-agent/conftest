#!/usr/bin/env bats

@test "Can verify policies that rely on annotations" {
  run $CONFTEST verify --data exclusions service.yaml

  [ "$status" -eq 0 ]
  echo $output
  [[ "$output" =~ "1 test, 1 passed, 0 warnings, 0 failures" ]]
}

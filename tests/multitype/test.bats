#!/usr/bin/env bats

@test "Can parse multi-type files" {
  run $CONFTEST test deployment.yaml grafana.ini
  [ "$status" -eq  1 ]
  [[ "$output" =~ "Port should be" ]]
}

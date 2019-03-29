#!/usr/bin/env bats

@test "Fail when testing an invalid service" {
  run conftest testdata/service.yaml
  [ "$status" -eq 1 ]
}

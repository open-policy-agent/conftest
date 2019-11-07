#!/usr/bin/env bats

@test "Can change output format in verify command" {
  run ./conftest verify -p examples/kubernetes/policy/ -o tap
  [[ "$output" =~ "ok" ]]
}

@test "Can verify rego tests" {
  run ./conftest verify --policy ./examples/kubernetes/policy
  [ "$status" -eq 0 ]
  [[ "$output" =~ "test_services_not_denied" ]]
}
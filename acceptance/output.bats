#!/usr/bin/env bats

@test "Can disable color" {
  run ./conftest test -p examples/kubernetes/policy examples/kubernetes/service.yaml --no-color
  [ "$status" -eq 0 ]
  [[ "$output" != *"[33m"* ]]
}

@test "Can change output format in test command" {
  run ./conftest test -p examples/kubernetes/policy/ -o tap examples/kubernetes/deployment.yaml
  [[ "$output" =~ "not ok" ]]
}

@test "Output results only once" {
  run ./conftest test -p examples/kubernetes/policy examples/kubernetes/deployment.yaml
  count="${#lines[@]}"
  [ "$count" -eq 3 ]
}
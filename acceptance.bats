#!/usr/bin/env bats

@test "Not fail when testing a service with a warning" {
  run conftest test -p examples/kubernetes/policy examples/kubernetes/service.yaml
  [ "$status" -eq 0 ]
}

@test "Fail when testing a deployment with root containers" {
  run conftest test -p examples/kubernetes/policy examples/kubernetes/deployment.yaml
  [ "$status" -eq 1 ]
}

@test "Fail when testing a service with warnings" {
  run conftest test --fail-on-warn -p examples/kubernetes/policy examples/kubernetes/service.yaml
  [ "$status" -eq 1 ]
}

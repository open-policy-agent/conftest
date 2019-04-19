#!/usr/bin/env bats

@test "Not fail when testing a service with a warning" {
  run conftest -p examples/kubernetes/policy examples/kubernetes/service.yaml
  [ "$status" -eq 0 ]
}

@test "Fail when testing a deployment with root containers" {
  run conftest -p examples/kubernetes/policy examples/kubernetes/deployment.yaml
  [ "$status" -eq 1 ]
}

@test "Fail when testing a service with warnings" {
  run conftest --fail-on-warn -p examples/kubernetes/policy examples/kubernetes/service.yaml
  [ "$status" -eq 1 ]
}

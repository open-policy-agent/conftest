#!/usr/bin/env bats

@test "Not fail when testing a service with a warning" {
  run ./conftest test -p examples/kubernetes/policy examples/kubernetes/service.yaml
  [ "$status" -eq 0 ]
}

@test "Fail when testing a deployment with root containers" {
  run ./conftest test -p examples/kubernetes/policy examples/kubernetes/deployment.yaml
  [ "$status" -eq 1 ]
}

@test "Fail when testing a service with warnings" {
  run ./conftest test --fail-on-warn -p examples/kubernetes/policy examples/kubernetes/service.yaml
  [ "$status" -eq 1 ]
}

@test "Pass when testing a blank namespace" {
  run ./conftest test --namespace notpresent -p examples/kubernetes/policy examples/kubernetes/deployment.yaml
  [ "$status" -eq 0 ]
}

@test "Fail due to picking up settings from configuration file" {
  cd examples/configfile
  run ../../conftest test deployment.yaml
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Containers must not run as root" ]]
}

@test "Has version flag" {
  run ./conftest --version
  [ "$status" -eq 0 ]
}

@test "Has trace flag" {
  run ./conftest test -p examples/kubernetes/policy examples/kubernetes/service.yaml --trace
  [ "$status" -eq 0 ]
  [[ "$output" =~ "Enter warn[msg] { data.kubernetes.is_service;" ]]
}

@test "Has help flag" {
  run ./conftest --help
  [ "$status" -eq 0 ]
}

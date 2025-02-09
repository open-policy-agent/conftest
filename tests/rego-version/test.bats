#!/usr/bin/env bats

@test "TEST - V0 Policy with V1 flag disabled" {
  run $CONFTEST test --policy=policy/v0.rego --rego-version=v0 data.yaml

  [ "$status" -eq 0 ]
  echo $output
}

@test "TEST - V0 Policy with V1 flag enabled" {
  run $CONFTEST test --policy=policy/v0.rego --rego-version=v1 data.yaml

  [ "$status" -eq 1 ]
  echo $output
}

@test "TEST - V1 Policy with V1 flag disabled" {
  run $CONFTEST test --policy=policy/v1.rego --rego-version=v0 data.yaml

  [ "$status" -eq 1 ]
  echo $output
}

@test "TEST - V1 Policy with V1 flag enabled" {
  run $CONFTEST test --policy=policy/v1.rego --rego-version=v1 data.yaml

  [ "$status" -eq 0 ]
  echo $output
}

@test "VERIFY - V0 Policy with V1 flag disabled" {
  run $CONFTEST verify --policy=policy/v0.rego --rego-version=v0

  [ "$status" -eq 0 ]
  echo $output
}

@test "VERIFY - V0 Policy with V1 flag enabled" {
  run $CONFTEST verify --policy=policy/v0.rego --rego-version=v1

  [ "$status" -eq 1 ]
  echo $output
}

@test "VERIFY - V1 Policy with V1 flag disabled" {
  run $CONFTEST verify --policy=policy/v1.rego --rego-version=v0

  [ "$status" -eq 1 ]
  echo $output
}

@test "VERIFY - V1 Policy with V1 flag enabled" {
  run $CONFTEST verify --policy=policy/v1.rego --rego-version=v1

  [ "$status" -eq 0 ]
  echo $output
}

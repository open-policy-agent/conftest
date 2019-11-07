#!/usr/bin/env bats

@test "Can parse tf files" {
  run ./conftest test -p examples/terraform/policy/gke.rego examples/terraform/gke.tf
  [ "$status" -eq 0 ]
}

@test "Can parse toml files" {
  run ./conftest test -p examples/traefik/policy examples/traefik/traefik.toml
  [ "$status" -eq 1 ]
}

@test "Can parse cue files" {
  run ./conftest test -p examples/cue/policy examples/cue/deployment.cue
  [ "$status" -eq 1 ]
  [[ "$output" =~ "The image port should be 8080 in deployment.cue. you got : 8081" ]]
}

@test "Can parse ini files" {
  run ./conftest test -p examples/ini/policy examples/ini/grafana.ini
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Users should verify their e-mail address" ]]
}

@test "Can parse hcl2 files" {
  run ./conftest test -p examples/hcl2/policy examples/hcl2/terraform.tf -i hcl2
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Application environment is should be `staging_environment`" ]]
}

@test "Can parse docker files" {
  run ./conftest test -p examples/docker/policy examples/docker/Dockerfile
  [ "$status" -eq 1 ]
  [[ "$output" =~ "blacklisted image found [\"openjdk:8-jdk-alpine\"]" ]]
}
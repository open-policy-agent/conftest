#!/usr/bin/env bats

@test "Can parse and print structured output from given input" {
  run ./conftest parse examples/docker/Dockerfile
  [ "$status" -eq 0 ]
  [ "${lines[0]}" = "examples/docker/Dockerfile" ]
  [[ "$output" =~ "openjdk:8-jdk-alpine" ]]
}
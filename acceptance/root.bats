#!/usr/bin/env bats

@test "Has version flag" {
  run ./conftest --version
  [ "$status" -eq 0 ]
}

@test "Has help flag" {
  run ./conftest --help
  [ "$status" -eq 0 ]
}
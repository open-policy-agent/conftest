#!/usr/bin/env bats

@test "Can document the policies" {
  rm "policy.md"
  run $CONFTEST doc ./policy

  [ "$status" -eq 0 ]
  echo $output
  [ -f "policy.md" ]
}

@test "Can document the sub package" {
  rm "sub.md"
  run $CONFTEST doc ./policy/sub

  [ "$status" -eq 0 ]
  echo $output
  [ -f "sub.md" ]
}

#!/usr/bin/env bats

@test "Can document the policies" {
  rm -f "policy.md"
  run $CONFTEST doc ./policy

  [ "$status" -eq 0 ]
  echo $output
  [ -f "policy.md" ]
}

@test "Can document the sub package" {
  rm -f "sub.md"
  run $CONFTEST doc ./policy/sub

  [ "$status" -eq 0 ]
  echo $output
  [ -f "sub.md" ]
}

@test "Can document using custom template and output" {
  rm -f "custom/policy.md"
  mkdir -p "custom"
  run $CONFTEST doc -t ./template.md.tpl -o ./custom ./policy

  [ "$status" -eq 0 ]
  echo $output
  [ -f "custom/policy.md" ]
}


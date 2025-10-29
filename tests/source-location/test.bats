#!/usr/bin/env bats

@test "Location is included in results" {
  run $CONFTEST test -o json data.yaml

  echo $output
  [[ "$output" =~ "\"file\": \"test.txt\"" ]]
  [[ "$output" =~ "\"line\": 123" ]]
}

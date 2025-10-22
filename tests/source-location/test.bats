#!/usr/bin/env bats

@test "Location is included in JSON results" {
  run $CONFTEST test -o json -d loc_different_file.yaml data.yaml

  echo $output
  [[ "$output" =~ "\"file\": \"test.txt\"" ]]
  [[ "$output" =~ "\"line\": 123" ]]
}

@test "Location is included in GitHub results - Different file" {
  run $CONFTEST test -o github -d loc_different_file.yaml data.yaml

  echo $output
  [[ "$output" =~ "::error file=test.txt,line=123::" ]]
  [[ "$output" =~ "::error file=data.yaml,line=1::(ORIGINATING FROM test.txt L123)" ]]
}

@test "Location is included in GitHub results - Same file" {
  run $CONFTEST test -o github -d loc_same_file.yaml data.yaml

  echo $output
  [[ "$output" =~ "::error file=data.yaml,line=123::" ]]
  [[ ! "$output" =~ "test.txt" ]]
}

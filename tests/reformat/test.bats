#!/usr/bin/env bats

DIR="$( cd "$( dirname "${BATS_TEST_FILENAME}" )" >/dev/null 2>&1 && pwd )"
PROJECT_ROOT="$( cd "$DIR/../.." >/dev/null 2>&1 && pwd )"

setup_file() {
    cd "$PROJECT_ROOT"

    # Generate test JSON data by running conftest test (ignore exit status since policy may fail)
    $CONFTEST test --output json -p examples/kubernetes/policy examples/kubernetes/deployment.yaml > "$DIR/test_results.json" || true
}

teardown_file() {
    # Clean up generated test file
    rm -f "$DIR/test_results.json"
    rm -f "$DIR/empty.json"
}

@test "Reformat JSON to table format using positional argument" {
  run $CONFTEST reformat "$DIR/test_results.json" --output table
  [ "$status" -eq 0 ]
  [[ "$output" =~ "RESULT" ]]
  [[ "$output" =~ "FILE" ]]
  [[ "$output" =~ "examples/kubernetes/deployment.yaml" ]]
}

@test "Reformat JSON to junit format using positional argument" {
  run $CONFTEST reformat "$DIR/test_results.json" --output junit
  [ "$status" -eq 0 ]
  [[ "$output" =~ "<?xml version=\"1.0\" encoding=\"UTF-8\"?>" ]]
  [[ "$output" =~ "<testsuites>" ]]
  [[ "$output" =~ "hello-kubernetes" ]]
}


@test "Reformat JSON via stdin to table format" {
  run bash -c "cat \"$DIR/test_results.json\" | $CONFTEST reformat --output table"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "RESULT" ]]
  [[ "$output" =~ "FILE" ]]
  [[ "$output" =~ "examples/kubernetes/deployment.yaml" ]]
}

@test "Reformat JSON via stdin to json format (default)" {
  run bash -c "cat \"$DIR/test_results.json\" | $CONFTEST reformat"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "examples/kubernetes/deployment.yaml" ]]
  [[ "$output" =~ "hello-kubernetes" ]]
}

@test "Reformat JSON via stdin to junit format" {
  run bash -c "cat \"$DIR/test_results.json\" | $CONFTEST reformat --output junit"
  [ "$status" -eq 0 ]
  [[ "$output" =~ "<?xml version=\"1.0\" encoding=\"UTF-8\"?>" ]]
  [[ "$output" =~ "<testsuites>" ]]
  [[ "$output" =~ "hello-kubernetes" ]]
}

@test "Handle empty stdin gracefully" {
  run bash -c "echo '' | $CONFTEST reformat --output table"
  [ "$status" -ne 0 ]
  [[ "$output" =~ "failed to parse JSON input" ]]
}

@test "Handle malformed JSON via stdin" {
  run bash -c "echo 'invalid json' | $CONFTEST reformat --output table"
  [ "$status" -ne 0 ]
  [[ "$output" =~ "failed to parse JSON input" ]]
}

@test "Fail when input file does not exist" {
  run $CONFTEST reformat nonexistent.json --output table
  [ "$status" -eq 1 ]
  [[ "$output" =~ "failed to open input file" ]]
}

@test "Fail when invalid JSON provided" {
  echo "invalid json" > invalid.json
  run $CONFTEST reformat invalid.json --output table
  [ "$status" -eq 1 ]
  [[ "$output" =~ "failed to parse JSON input" ]]
  rm -f invalid.json
}

@test "Handle invalid output format gracefully" {
  run $CONFTEST reformat "$DIR/test_results.json" --output invalidformat
  [ "$status" -eq 0 ]
  # Invalid format defaults to standard output format
  [[ "$output" =~ "hello-kubernetes" ]]
}

@test "Handle empty JSON array" {
  echo '[]' > "$DIR/empty.json"
  run $CONFTEST reformat "$DIR/empty.json" --output table
  [ "$status" -eq 0 ]
  # Empty array should produce empty table output
  [[ ! "$output" =~ "hello-kubernetes" ]]
}

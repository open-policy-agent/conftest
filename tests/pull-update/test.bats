#!/usr/bin/env bats

setup_file() {
	# Create a temporary directory shared by all the tests
	export TEMP_DIR=$(mktemp -d)

	# Copy all the files there
	cp -r . "${TEMP_DIR}"
}

teardown_file() {
	# Cleanup temporary directory
	rm -rf "${TEMP_DIR}"
}

@test "Ensure that policy do not exists" {
	cd "${TEMP_DIR}"
	run test -e "policy"

	[ "$status" -eq 1 ]
}

@test "Pull and update first version policy" {
	cd "${TEMP_DIR}"
	run $CONFTEST test --update "file://${PWD}/remote-policy/a" file.json

	[ "$status" -eq 1 ]
	[[ "$output" =~ "a should not be present" ]]
	[[ "$output" =~ "1 test, 0 passed, 0 warnings, 1 failure, 0 exceptions" ]]
}

@test "Ensure that policy directory exists" {
	cd "${TEMP_DIR}"
	run test -d "policy"

	[ "$status" -eq 0 ]
}

@test "Pull and update second version policy" {
	cd "${TEMP_DIR}"
	run $CONFTEST test --update "file://${PWD}/remote-policy/b" file.json

	[ "$status" -eq 1 ]
	[[ "$output" =~ "a should not be present" ]]
	[[ "$output" =~ "b should not be present" ]]
	[[ "$output" =~ "2 tests, 0 passed, 0 warnings, 2 failures, 0 exceptions" ]]
}

#!/usr/bin/env bats

setup_file() {
	# Create a temporary directory shared by all the tests
	export TEMP_DIR=$(mktemp -d)

	# Copy all the files there
	cp -r . "${TEMP_DIR}"

	# On Windows (MSYS2/Git Bash), convert to mixed-style path for conftest compatibility
	if command -v cygpath >/dev/null 2>&1; then
		# Convert and explicitly re-export the converted path
		export TEMP_DIR=$(cygpath -m "${TEMP_DIR}")
	fi

	# Debug output for CI troubleshooting (visible in bats output)
	echo "# DEBUG: TEMP_DIR=${TEMP_DIR}" >&3
	echo "# DEBUG: Contents of TEMP_DIR:" >&3
	ls -la "${TEMP_DIR}" >&3 2>&1 || true
}

teardown_file() {
	# Cleanup temporary directory
	rm -rf "${TEMP_DIR}"
}

@test "Ensure that policy do not exists" {
	run test -e "${TEMP_DIR}/policy"

	[ "$status" -eq 1 ]
}

@test "Pull and update first version policy" {
	# Debug: Show what we're about to run
	echo "# DEBUG: Running conftest with TEMP_DIR=${TEMP_DIR}" >&3
	echo "# DEBUG: Update URL=file:///${TEMP_DIR}/remote-policy/a" >&3

	run $CONFTEST test --policy "${TEMP_DIR}/policy" --update "file:///${TEMP_DIR}/remote-policy/a" "${TEMP_DIR}/file.json"

	# Debug: Show actual output for troubleshooting
	echo "# DEBUG: Exit status=$status" >&3
	echo "# DEBUG: Output=$output" >&3

	[ "$status" -eq 1 ]
	[[ "$output" =~ "a should not be present" ]]
	[[ "$output" =~ "1 test, 0 passed, 0 warnings, 1 failure, 0 exceptions" ]]
}

@test "Ensure that policy directory exists" {
	run test -d "${TEMP_DIR}/policy"

	[ "$status" -eq 0 ]
}

@test "Pull and update second version policy" {
	# Debug: Show what we're about to run
	echo "# DEBUG: Running conftest with TEMP_DIR=${TEMP_DIR}" >&3
	echo "# DEBUG: Update URL=file:///${TEMP_DIR}/remote-policy/b" >&3

	run $CONFTEST test --policy "${TEMP_DIR}/policy" --update "file:///${TEMP_DIR}/remote-policy/b" "${TEMP_DIR}/file.json"

	# Debug: Show actual output for troubleshooting
	echo "# DEBUG: Exit status=$status" >&3
	echo "# DEBUG: Output=$output" >&3

	[ "$status" -eq 1 ]
	[[ "$output" =~ "a should not be present" ]]
	[[ "$output" =~ "b should not be present" ]]
	[[ "$output" =~ "2 tests, 0 passed, 0 warnings, 2 failures, 0 exceptions" ]]
}

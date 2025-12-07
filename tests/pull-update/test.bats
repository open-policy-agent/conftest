#!/usr/bin/env bats

setup_file() {
	# Create a temporary directory shared by all the tests
	export TEMP_DIR=$(mktemp -d)

	# Copy all the files there
	cp -r . "${TEMP_DIR}"

	# On Windows (MSYS2/Git Bash), paths need to be converted for native executables.
	if command -v cygpath >/dev/null 2>&1; then
		export TEMP_DIR_WIN=$(cygpath -m "${TEMP_DIR}")
	else
		export TEMP_DIR_WIN="${TEMP_DIR}"
	fi

	# Create a git repository for the remote-policy to enable git:// URL downloads
	# This avoids the file:// symlink issues on Windows
	cd "${TEMP_DIR}/remote-policy"
	git init --quiet
	git config user.email "test@test.com"
	git config user.name "Test"
	git add .
	git commit --quiet -m "initial"
	cd - >/dev/null
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
	# Use git:: prefix for local git repository to avoid file:// symlink issues on Windows
	run $CONFTEST test --policy "${TEMP_DIR_WIN}/policy" --update "git::file://${TEMP_DIR_WIN}/remote-policy//a" "${TEMP_DIR_WIN}/file.json"

	[ "$status" -eq 1 ]
	[[ "$output" =~ "a should not be present" ]]
	[[ "$output" =~ "1 test, 0 passed, 0 warnings, 1 failure, 0 exceptions" ]]
}

@test "Ensure that policy directory exists" {
	run test -d "${TEMP_DIR}/policy"

	[ "$status" -eq 0 ]
}

@test "Pull and update second version policy" {
	# Use git:: prefix for local git repository to avoid file:// symlink issues on Windows
	run $CONFTEST test --policy "${TEMP_DIR_WIN}/policy" --update "git::file://${TEMP_DIR_WIN}/remote-policy//b" "${TEMP_DIR_WIN}/file.json"

	[ "$status" -eq 1 ]
	[[ "$output" =~ "a should not be present" ]]
	[[ "$output" =~ "b should not be present" ]]
	[[ "$output" =~ "2 tests, 0 passed, 0 warnings, 2 failures, 0 exceptions" ]]
}

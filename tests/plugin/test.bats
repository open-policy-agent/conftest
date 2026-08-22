setup() {
  # kubectl-conftest.sh calls check_command for both conftest and kubectl and
  # exits 1 if either is missing. The binary under test is invoked through
  # $CONFTEST rather than the PATH, so the plugin never found it and always
  # exited 1 here; that used to be reported as success because Exec mapped a
  # plugin exit code of 1 to nil. Put both on the PATH so these tests exercise
  # the plugin rather than its precondition check.
  local conftest_dir stub_dir
  conftest_dir="$(cd "$(dirname "$CONFTEST")" >/dev/null 2>&1 && pwd)"
  stub_dir="${BATS_TEST_TMPDIR:-$BATS_TMPDIR}/bin"

  mkdir -p "$stub_dir"
  # the tests below run the plugin with no arguments, which only prints usage,
  # so kubectl needs to exist but is never actually called
  printf '#!/usr/bin/env bash\nexit 0\n' >"$stub_dir/kubectl"
  chmod +x "$stub_dir/kubectl"

  PATH="$conftest_dir:$stub_dir:$PATH"
}

@test "Can install plugin from directory" {
  run $CONFTEST plugin install ../../contrib/plugins/kubectl
  [ "$status" -eq 0 ]

  run $CONFTEST kubectl
  [ "$status" -eq 0 ]
}

@test "Can install plugin from URL" {
  run $CONFTEST plugin install github.com/open-policy-agent/conftest/contrib/plugins/kubectl
  [ "$status" -eq 0 ]

  run $CONFTEST kubectl
  [ "$status" -eq 0 ]
}

@test "Plugin exit code is propagated" {
  run $CONFTEST plugin install ../../contrib/plugins/echo
  [ "$status" -eq 0 ]

  run $CONFTEST echo hello
  [ "$status" -eq 0 ]
  [ "$output" = "hello" ]

  run $CONFTEST echo 42 some message
  [ "$status" -eq 42 ]
}

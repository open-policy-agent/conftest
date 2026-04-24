@test "Can install plugin from directory" {
  run $CONFTEST plugin install ../../contrib/plugins/echo
  [ "$status" -eq 0 ]

  run $CONFTEST echo hello
  [ "$status" -eq 0 ]
  [ "$output" = "hello" ]
}

@test "Can install plugin from URL" {
  run $CONFTEST plugin install github.com/open-policy-agent/conftest/contrib/plugins/kubectl
  [ "$status" -eq 0 ]
}

@test "Plugin exit code is propagated" {
  run $CONFTEST plugin install ../../contrib/plugins/echo
  [ "$status" -eq 0 ]

  run $CONFTEST echo 42 some message
  [ "$status" -eq 42 ]
}

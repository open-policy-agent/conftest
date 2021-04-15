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

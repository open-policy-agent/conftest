package main
import rego.v1

deny_valid_uri contains msg if {
  some name
  value := input[name]
  contains(lower(name), "url")
  not contains(lower(value), "http")
  msg := sprintf("Must have a valid uri defined '%s'", [value])
}

secret_exceptions := {
 "secret.value.exception"
}

deny_no_secrets contains msg if {
  some name
  value := input[name]
  not secret_exceptions[name]
  contains(lower(name), "secret")
  msg := sprintf("'%s' may contain a secret value", [name])
}

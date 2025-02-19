package main
import rego.v1

test_deny_valid if {
    not deny with input as parse_config_file("file_does_not_exist.yaml")
}

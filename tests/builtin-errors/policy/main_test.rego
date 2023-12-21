package main

test_deny_valid {
    not deny with input as parse_config_file("file_does_not_exist.yaml")
}

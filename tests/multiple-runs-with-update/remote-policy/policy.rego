package main

import rego.v1

deny contains msg if {
    input.bar == "baz"
    msg := "remote-policy"
}

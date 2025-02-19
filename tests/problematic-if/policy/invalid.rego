package main
import rego.v1


deny[msg] if {
    msg := "foo"
}

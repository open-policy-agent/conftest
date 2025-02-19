package main
import rego.v1


deny contains msg if {
    msg := "foo"
}

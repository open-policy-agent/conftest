package main
import rego.v1

deny contains msg if {
	input.foo = "bar"
	msg = "Org policy forbids foo=bar"
}

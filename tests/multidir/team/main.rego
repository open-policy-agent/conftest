package main
import rego.v1

deny contains msg if {
	input.baz = "qux"
	msg = "Team policy forbids baz=qux"
}

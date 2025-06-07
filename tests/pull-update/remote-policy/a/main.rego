package main

import rego.v1

deny contains msg if {
	input.a
	msg := "a should not be present"
}

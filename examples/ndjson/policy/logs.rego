package main
import rego.v1

deny contains msg if {
	input.level == "ERROR"
	msg = sprintf("error log found, message '%s'", [input.message])
}

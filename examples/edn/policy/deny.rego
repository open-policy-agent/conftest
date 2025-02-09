package main
import rego.v1

deny contains msg if {
	input[":env"] = ":development"
	input[":log"] != ":debug"
	msg = "Applications in the development environment should have debug logging"
}

deny contains msg if {
	input[":env"] = ":production"
	input[":log"] != ":error"
	msg = "Applications in the production environment should have error only logging"
}

package main
import rego.v1

version := to_number(input.version)

deny contains msg if {
	endswith(input.services[_].image, ":latest")
	msg = "No images tagged latest"
}

deny contains msg if {
	version < 3.5
	msg = "Must be using at least version 3.5 of the Compose file format"
}

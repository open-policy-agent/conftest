package main
import rego.v1

failures = ["one", "two", "three"]

deny contains resource_name if {
	resource_name = failures[_]
}

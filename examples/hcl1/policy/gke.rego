package main
import rego.v1

deny contains sprintf("could not find any resources in: %v", [input]) if {
	count(input.resource) == 0
}

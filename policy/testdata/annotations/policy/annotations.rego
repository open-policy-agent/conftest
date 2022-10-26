package main

# METADATA
# title: annotations test
# description: Violation message pulled from metadata.
deny[result] {
	metadata := rego.metadata.rule()
	result := {"msg": metadata.description}
}

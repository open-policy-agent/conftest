package main

import data.services

name := input.metadata.name

kind := input.kind

type := input.spec.type

# METADATA
# title: Example using annotations
# custom:
#   template: 'Cannot expose port %v on LoadBalancer. Denied ports: %v'
deny[msg] {
	kind == "Service"
	type == "LoadBalancer"

	some p
	input.spec.ports[p].port

	input.spec.ports[p].port == services.ports[_]

	metadata := rego.metadata.rule()
	msg := sprintf(metadata.custom.template, [input.spec.ports[p].port, services.ports])
}

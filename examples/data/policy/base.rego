package main

import data.services

name := input.metadata.name

kind := input.kind

type := input.spec.type

deny[msg] {
	kind == "Service"
	type == "LoadBalancer"

	some p
	input.spec.ports[p].port

	input.spec.ports[p].port == services.ports[_]

	msg := sprintf("Cannot expose port %v on LoadBalancer. Denied ports: %v", [input.spec.ports[p].port, services.ports])
}

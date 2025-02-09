# METADATA
# title: Example using annotations
# description: This package validates that ...
# custom:
#   template: 'Cannot expose port %v on LoadBalancer. Denied ports: %v'
package main.sub
import rego.v1

import data.services

name := input.metadata.name
kind := input.kind
type := input.spec.type

# METADATA
# title: Example using annotations
# description: This rule validates that ...
# custom:
#   template: 'Cannot expose port %v on LoadBalancer. Denied ports: %v'
deny contains msg if {
	kind == "Service"
	type == "LoadBalancer"

	some p
	input.spec.ports[p].port

	input.spec.ports[p].port == services.ports[_]

	metadata := rego.metadata.rule()
	msg := sprintf(metadata.custom.template, [input.spec.ports[p].port, services.ports])
}

# METADATA
# title: Second Example using annotations
# description: This rule validates that ...
# custom:
#   template: 'Cannot expose port %v on LoadBalancer. Denied ports: %v'
deny contains msg if {
	kind == "Service"
	type == "LoadBalancer"

	some p
	input.spec.ports[p].port

	input.spec.ports[p].port == services.ports[_]

	metadata := rego.metadata.rule()
	msg := sprintf(metadata.custom.template, [input.spec.ports[p].port, services.ports])
}


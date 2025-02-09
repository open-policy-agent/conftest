package main
import rego.v1

import data.kubernetes

name := input.metadata.name

warn contains msg if {
	kubernetes.is_service
	msg = sprintf("Found service %s but services are not allowed", [name])
}

package main
import rego.v1

is_deployment if {
	input.kind = "Deployment"
}

deny_run_as_root contains msg if {
	is_deployment
	not input.spec.template.spec.securityContext.runAsNonRoot

	msg = sprintf("Containers must not run as root in Deployment %s", [input.metadata.name])
}

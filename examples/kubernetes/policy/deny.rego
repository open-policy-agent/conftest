package main
import rego.v1

import data.kubernetes

name := input.metadata.name

deny contains msg if {
	kubernetes.is_deployment
	not input.spec.template.spec.securityContext.runAsNonRoot

	msg = sprintf("Containers must not run as root in Deployment %s", [name])
}

required_deployment_selectors if {
	input.spec.selector.matchLabels.app
	input.spec.selector.matchLabels.release
}

deny contains msg if {
	kubernetes.is_deployment
	not required_deployment_selectors

	msg = sprintf("Deployment %s must provide app/release labels for pod selectors", [name])
}

package main
import rego.v1

deny contains msg if {
	input.kind = "Deployment"
	not input.spec.template.spec.securityContext.runAsNonRoot = true
	msg = "Containers must not run as root"
}

deny contains msg if {
	input.kind = "Deployment"
	not input.spec.selector.matchLabels.app
	msg = "Containers must provide app label for pod selectors"
}

warn contains msg if {
	input.kind = "Service"
	msg = "Services are not allowed"
}

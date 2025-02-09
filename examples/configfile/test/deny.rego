package conftest
import rego.v1

import data.kubernetes

deny contains msg if {
	kubernetes.is_deployment
	not input.spec.template.spec.securityContext.runAsNonRoot = true
	msg = "Containers must not run as root"
}

deny contains msg if {
	kubernetes.is_deployment
	not input.spec.selector.matchLabels.app
	msg = "Containers must provide app label for pod selectors"
}

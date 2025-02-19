package main
import rego.v1

name := input.metadata.name

required_deployment_labels if {
	input.metadata.labels["app.kubernetes.io/name"]
	input.metadata.labels["app.kubernetes.io/instance"]
}

deny contains msg if {
	input.kind == "Deployment"
	# regal ignore:print-or-trace-call
	print(name)
	not required_deployment_labels
	msg := sprintf("%s must include Kubernetes recommended labels: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels", [name])
}

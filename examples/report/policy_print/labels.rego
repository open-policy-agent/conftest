package main

name = input.metadata.name

required_deployment_labels {
	input.metadata.labels["app.kubernetes.io/name"]
	input.metadata.labels["app.kubernetes.io/instance"]
}

deny[msg] {
	input.kind = "Deployment"
	print(name)
	not required_deployment_labels
	msg := sprintf("%s must include Kubernetes recommended labels: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels", [name])
}

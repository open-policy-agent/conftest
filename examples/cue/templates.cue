package kubernetes

deployment <Name>: {
	apiVersion: string
	kind:       "Deployment"
	metadata name: Name
	spec: {
		replicas: 1 | int
		selector matchLabels app: Name
		template: {
			metadata labels app: Name
			spec containers: [{name: Name}]
			// spec securityContext runAsNonRoot: true
		}
	}
}

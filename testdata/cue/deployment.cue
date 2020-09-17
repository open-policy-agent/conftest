package kubernetes


deployment "hello-kubernetes": {
	apiVersion: "apps/v1"
	spec: {
		replicas: 3
		template spec containers: [{
			image: "paulbouwer/hello-kubernetes:1.5"
			ports: [{
				containerPort: 8081
			}]
		}]
	}
}

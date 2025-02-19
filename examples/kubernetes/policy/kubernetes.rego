package kubernetes
import rego.v1

is_service if {
	input.kind = "Service"
}

is_deployment if {
	input.kind = "Deployment"
}

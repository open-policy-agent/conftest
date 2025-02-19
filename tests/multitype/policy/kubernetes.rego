package kubernetes
import rego.v1

is_deployment if {
	input.kind = "Deployment"
}

package main
import rego.v1

deny contains msg if {
	not input.apiVersion = "apps/v1"
	msg = sprintf("apiVersion must be apps/v1 in : %s", [input.metadata.name])
}

deny contains msg if {
	repl := input.spec.replicas
	repl < 2
	msg = sprintf("Replica count must be greater than 2, you have : %d", [repl])
}

deny contains msg if {
	ports := input.spec.template.spec.containers[_].ports[_].containerPort
	not ports = 8080
	msg = sprintf("The image port should be 8080 in deployment.cue. you have : %d", [ports])
}

deny contains msg if {
	endswith(input.spec.template.spec.containers[_].image, ":latest")
	msg = "No images tagged latest"
}

package main

deny[msg] {
	not input.apiVersion = "apps/v1"
	msg = sprintf("apiVersion must be apps/v1 in : %s", [input.metadata.name])
}

deny[msg] {
	repl := input.spec.replicas
	repl < 2
	msg = sprintf("Replica count must be greater than 2, you have : %d", [repl])
}

deny[msg] {
	ports := input.spec.template.spec.containers[_].ports[_].containerPort
	not ports = 8080
	msg = sprintf("The image port should be 8080 in deployment.cue. you have : %d", [ports])
}

deny[msg] {
	endswith(input.spec.template.spec.containers[_].image, ":latest")
	msg = "No images tagged latest"
}

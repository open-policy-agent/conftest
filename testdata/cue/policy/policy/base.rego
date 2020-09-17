package main

name = input.deployment[_]

deny[msg] {
  not name.apiVersion = "apps/v1"
  msg = sprintf("Api Version must be apps/v1 in : %s", [name])
}

deny[msg] {
  repl := name.spec.replicas
  repl < 3
  msg = sprintf("Replica count must be higher than 3, you have : %d", [repl])
}

deny[msg] {
  ports := name.spec.template.spec.containers[_].ports[_].containerPort
  not ports = 8080
  msg = sprintf("The image port should be 8080 in deployment.cue. you got : %d", [ports])
}

deny[msg] {
  endswith(name.spec.template.spec.containers[_].image, ":latest")
  msg = "No images tagged latest"
}
package main

# Prevent ingress from pointing at service ports which aren't defined (only applies if service and ingress are both part of input, but can be in separate files)
deny[msg] {
  ingress := input[ingressFile][_] # With --combine and a multi-document yaml, input[filename] is always an array of subdocuments
  ingress.kind == "Ingress" # Technically we should also check the apiVersion

  serviceName := ingress.spec.rules[ruleIndex].http.paths[pathIndex].backend.serviceName
  servicePort := ingress.spec.rules[ruleIndex].http.paths[pathIndex].backend.servicePort

  service := input[serviceFile][_]
  service.kind == "Service"

  service.metadata.name == serviceName # Now `service` should be a service pointed at by `ingress`
  matchedPorts := [port.name | port = service.spec.ports[_]; port.name == servicePort] # List should be either length 0 (invalid port name) or 1 (valid port name)

  count(matchedPorts) == 0

  msg := sprintf("Ingress '%s' (in %s) points at port '%s' in service '%s' (in %s). However this service doesn't define this port (available ports: %s)", [
    ingress.metadata.name, ingressFile, servicePort, serviceName, serviceFile, concat(", ", [port.name | port = service.spec.ports[_]])
  ])
}

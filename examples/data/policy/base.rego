package main

import data.services

name = input.metadata.name
kind = input.kind
type = input.spec.type

deny[msg] {
  kind = "Service"
  type = "LoadBalancer"
  input.spec.ports[_].port = services.ports[_]

  msg = sprintf("Cannot expose one of the following ports on a LoadBalancer %s", [services.ports])
}

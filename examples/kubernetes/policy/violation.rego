package main

import data.kubernetes

name = input.metadata.name

violation[{"msg": msg, "details": {}}] {
  kubernetes.is_service
  input.spec.type == "LoadBalancer"
  msg = sprintf("Service %s has type LoadBalancer which is not allowed", [name])
}

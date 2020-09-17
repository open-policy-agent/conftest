package main

is_deployment {
  input.kind = "Deployment"
}

deny_run_as_root[msg] {
  is_deployment
  not input.spec.template.spec.securityContext.runAsNonRoot

  msg = sprintf("Containers must not run as root in Deployment %s", [input.metadata.name])
}

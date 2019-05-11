package main

import data.kubernetes

name = input.metadata.name

deny[msg] {
  kubernetes.is_deployment
  not input.spec.template.spec.securityContext.runAsNonRoot = true
  msg = sprintf("Containers must not run as root in Deployment %s", [name])
}

deny[msg] {
  kubernetes.is_deployment
  not input.spec.selector.matchLabels.app
  not input.spec.selector.matchLabels.release
  msg = sprintf("Deployment %s must provide app/release labels for pod selectors", [name])
}

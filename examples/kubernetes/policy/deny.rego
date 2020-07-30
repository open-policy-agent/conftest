package main

import data.kubernetes

name = input.metadata.name

deny[msg] {
  kubernetes.is_deployment
  not input.spec.template.spec.securityContext.runAsNonRoot

  msg = sprintf("Containers must not run as root in Deployment %s", [name])
}

deny[msg] {
  input.kind == "Deployment"
  container := input.spec.template.spec.containers[_]
  not container.securityContext.allowPrivilegeEscalation == false
  msg = "Containers must not allow privilege escalation"
}

deny[msg] {
  input.kind == "Deployment"
  container := input.spec.template.spec.containers[_]
  not container.securityContext.capabilities.drop == ["AUDIT_WRITE", "CHOWN", "DAC_OVERRIDE", "FOWNER", "FSETID", "KILL", "MKNOD", "NET_BIND_SERVICE", "NET_RAW", "SETFCAP", "SETGID", "SETPCAP", "SETUID", "SYS_CHROOT"]
  msg = "Containers must drop following capabilities: AUDIT_WRITE, CHOWN, DAC_OVERRIDE, FOWNER, FSETID, KILL, MKNOD, NET_BIND_SERVICE, NET_RAW, SETFCAP, SETGID, SETPCAP, SETUID, SYS_CHROOT "
}

deny[msg] {
  input.kind == "Deployment"
  not input.spec.template.spec.automountServiceAccountToken == false
  msg = "Containers must not allow automount serviceaccount token"
}

required_deployment_selectors {
  input.spec.selector.matchLabels.app
  input.spec.selector.matchLabels.release
}

deny[msg] {
  kubernetes.is_deployment
  not required_deployment_selectors

  msg = sprintf("Deployment %s must provide app/release labels for pod selectors", [name])
}

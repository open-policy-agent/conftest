package main

import data.kubernetes

name = input.metadata.name

violation[{"msg": msg, "details": {}}] {
	kubernetes.is_deployment
	msg = sprintf("Found deployment %s but deployments are not allowed", [name])
}

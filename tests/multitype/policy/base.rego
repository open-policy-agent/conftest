package main
import rego.v1

import data.grafana
import data.kubernetes

port := 3000

deny contains msg if {
	kubernetes.is_deployment
	not input.spec.template.spec.containers[0].ports[0].containerPort = 3000
	msg = sprintf("Port should be %d", [port])
}

deny contains msg if {
	grafana.is_config
	not input.server.http_port = "3000"
	msg = sprintf("Port should be %d", [port])
}

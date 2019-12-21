package main

import data.kubernetes
import data.grafana

port := 3000

deny[msg] {
    kubernetes.is_deployment
    not input.spec.template.spec.containers[0].ports[0].containerPort = 3000
    msg = sprintf("Port should be %d", [port])
}

deny[msg] {
    grafana.is_config
    not input.server.http_port = "3000"
    msg = sprintf("Port should be %d", [port])
}

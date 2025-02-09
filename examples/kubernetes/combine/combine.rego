package main
import rego.v1

violation := [msg] if {
    some i
    input[i].contents.kind == "Deployment"
    deployment := input[i].contents
    not service_selects_app(deployment.spec.selector.matchLabels.app)
    msg := sprintf("Deployment '%v' has no matching service", [deployment.metadata.name])
}

service_selects_app(app) if {
    some i
    input[i].contents.kind == "Service"
    service := input[i].contents
    service.spec.selector.app == app
}

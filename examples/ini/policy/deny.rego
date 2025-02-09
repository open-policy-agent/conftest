package main
import rego.v1

deny contains msg if {
	not input.alerting.enabled = true
	msg = "Alerting should turned on"
}

deny contains msg if {
	not input["auth.basic"].enabled = true
	msg = "Basic auth should be enabled"
}

deny contains msg if {
	not input.server.http_port = 3000
	msg = "Grafana port should be 3000"
}

deny contains msg if {
	not input.server.protocol = "http"
	msg = "Grafana should use default http"
}

deny contains msg if {
	not input.users.allow_sign_up = false
	msg = "Users cannot sign up themselves"
}

deny contains msg if {
	not input.users.verify_email_enabled = true
	msg = "Users should verify their e-mail address"
}

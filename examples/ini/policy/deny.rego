package main

deny[msg] {
  not input.alerting.enabled = true
  msg = "Alerting should turned on"
}

deny[msg] {
  not input["auth.basic"].enabled = true
  msg = "Basic auth should be enabled"
}

deny[msg] {
  not input.server.http_port = 3000
  msg = "Grafana port should be 3000"
}

deny[msg] {
  not input.server.protocol = "http"
  msg = "Grafana should use default http"
}

deny[msg] {
  not input.users.allow_sign_up = false
  msg = "Users cannot sign up themselves"
}

deny[msg] {
  not input.users.verify_email_enabled = true
  msg = "Users should verify their e-mail address"
}

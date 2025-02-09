package main
import rego.v1

deny contains msg if {
	not input.play.server.http.port = 9000
	msg = "Play http server port should be 9000"
}

deny contains msg if {
	not input.play.server.http.address = "0.0.0.0"
	msg = "Play http server bind address should be 0.0.0.0"
}

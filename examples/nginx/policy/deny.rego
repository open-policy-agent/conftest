package main

import rego.v1

# Collect all server blocks nested inside http blocks.
server_blocks contains server if {
	some directive in input.directives
	directive.name == "http"
	some server in directive.block.directives
	server.name == "server"
}

# A server listens without SSL if any listen directive does not include "ssl" as a parameter.
listens_without_ssl(server) if {
	some directive in server.block.directives
	directive.name == "listen"
	not "ssl" in directive.parameters
}

# A server redirects to HTTPS if it has a "return 301 https://..." directive.
redirects_to_https(server) if {
	some directive in server.block.directives
	directive.name == "return"
	directive.parameters[0] == "301"
	startswith(directive.parameters[1], "https://")
}

deny contains msg if {
	some server in server_blocks
	listens_without_ssl(server)
	not redirects_to_https(server)
	msg := "Server listening without SSL must redirect to HTTPS"
}

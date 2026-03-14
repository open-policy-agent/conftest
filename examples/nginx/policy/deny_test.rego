package main

import rego.v1

empty(value) if {
	count(value) == 0
}

no_violations if {
	empty(deny)
}

# A config with only an SSL listener should have no violations.
test_ssl_listener if {
	cfg := parse_config("nginx", `
		http {
			server {
				listen 443 ssl;
				server_name example.com;
			}
		}
	`)
	no_violations with input as cfg
}

# A server that redirects port 80 to HTTPS should have no violations.
test_port_80_redirects_to_https if {
	cfg := parse_config("nginx", `
		http {
			server {
				listen 80;
				server_name example.com;
				return 301 https://$host$request_uri;
			}
		}
	`)
	no_violations with input as cfg
}

# A server listening on port 80 without a redirect should be denied.
test_port_80_without_redirect if {
	cfg := parse_config("nginx", `
		http {
			server {
				listen 80;
				server_name example.com;
				location / {
					root /var/www/html;
				}
			}
		}
	`)
	deny["Server listening without SSL must redirect to HTTPS"] with input as cfg
}

# A non-HTTPS redirect (e.g. return 301 http://...) should still be denied.
test_port_80_redirects_to_http if {
	cfg := parse_config("nginx", `
		http {
			server {
				listen 80;
				server_name example.com;
				return 301 http://other.example.com;
			}
		}
	`)
	deny["Server listening without SSL must redirect to HTTPS"] with input as cfg
}

# A non-standard HTTP port (e.g. 8080) without a redirect should also be denied.
test_non_standard_port_without_redirect if {
	cfg := parse_config("nginx", `
		http {
			server {
				listen 8080;
				server_name example.com;
				location / {
					root /var/www/html;
				}
			}
		}
	`)
	deny["Server listening without SSL must redirect to HTTPS"] with input as cfg
}

# A non-standard HTTP port that redirects to HTTPS should have no violations.
test_non_standard_port_redirects_to_https if {
	cfg := parse_config("nginx", `
		http {
			server {
				listen 8080;
				server_name example.com;
				return 301 https://$host$request_uri;
			}
		}
	`)
	no_violations with input as cfg
}

# The sample nginx.conf uses a port-80-to-HTTPS redirect and should pass.
test_sample_config if {
	cfg := parse_config_file("../nginx.conf")
	no_violations with input as cfg
}

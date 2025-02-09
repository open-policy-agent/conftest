package grafana
import rego.v1

is_config if {
	input.server.protocol = http
}

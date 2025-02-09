package main
import rego.v1

test_parse_combined_config_file if {
	count(violation) == 1 with input as parse_combined_config_files(["combine.yaml"])
}

test_parse_combined_config_files if {
	count(violation) == 1 with input as parse_combined_config_files(["deployment.yaml", "service.yaml"])
}

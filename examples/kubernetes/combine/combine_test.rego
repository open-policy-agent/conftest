package main

test_parse_combined_config_file {
	count(violation) == 1 with input as parse_combined_config_files(["combine.yaml"])
}

test_parse_combined_config_files {
	count(violation) == 1 with input as parse_combined_config_files(["deployment.yaml", "service.yaml"])
}

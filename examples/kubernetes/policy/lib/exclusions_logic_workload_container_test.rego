package lib.exclusion_logic

workload_container_exclusion_samples = {"rule_names": {"no-capes": {"mr-incredible": {
	"shirt": "Who wears a cape with a shirt? Ridiculous.",
	"tie": "",
}}}}

test_excluded_when_workload_and_container_are_with_reason {
	rule_name := "no-capes"
	workload_name := "mr-incredible"
	container_name := "shirt"
	workload_container_is_excluded_from(rule_name, workload_name, container_name) with data.exclusions as workload_container_exclusion_samples
}

test_not_excluded_when_workload_and_container_are_but_no_reason_given {
	rule_name := "no-capes"
	workload_name := "mr-incredible"
	container_name := "tie"
	not workload_container_is_excluded_from(rule_name, workload_name, container_name) with data.exclusions as workload_container_exclusion_samples
}

test_not_excluded_when_workload_is_not {
	rule_name := "no-capes"
	workload_name := "elastigirl"
	container_name := "shirt"
	not workload_container_is_excluded_from(rule_name, workload_name, container_name) with data.exclusions as workload_container_exclusion_samples
}

test_not_excluded_when_workload_is_but_container_is_not {
	rule_name := "no-capes"
	workload_name := "mr-incredible"
	container_name := "socks"
	not workload_container_is_excluded_from(rule_name, workload_name, container_name) with data.exclusions as workload_container_exclusion_samples
}

test_not_excluded_when_rule_not_present {
	rule_name := "maintain-secret-identities"
	workload_name := "mr-incredible"
	container_name := "shirt"
	not workload_container_is_excluded_from(rule_name, workload_name, container_name) with data.exclusions as workload_container_exclusion_samples
}

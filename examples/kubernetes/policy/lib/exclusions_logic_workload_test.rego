package lib.exclusion_logic

workload_exclusion_samples = {"rule_names": {"no-capes": {"mr-incredible": {
}}}}

test_excluded_when_workload_is_present {
	rule_name := "no-capes"
	workload_name := "mr-incredible"
	workload_is_excluded_from(rule_name, workload_name) with data.exclusions as workload_exclusion_samples
}

test_not_excluded_when_workload_is_not_present {
	rule_name := "no-capes"
	workload_name := "elastigirl"
	not workload_is_excluded_from(rule_name, workload_name) with data.exclusions as workload_exclusion_samples
}

test_not_excluded_when_rule_is_not_present {
	rule_name := "maintain-secret-identities"
	workload_name := "mr-incredible"
	not workload_is_excluded_from(rule_name, workload_name) with data.exclusions as workload_exclusion_samples
}

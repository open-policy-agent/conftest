package lib.exclusion_logic

import data.exclusions
import data.lib.common

workload_is_excluded_from(rule, workload) {
	common.has_field(exclusions, "rule_names")
	common.has_field(exclusions.rule_names, rule)
	common.has_field(exclusions.rule_names[rule], workload)
}

workload_container_is_excluded_from(rule, workload, container) {
	common.has_field(exclusions, "rule_names")
	common.has_field(exclusions.rule_names, rule)
	common.has_field(exclusions.rule_names[rule], workload)
	common.has_field(exclusions.rule_names[rule][workload], container)
    # Must specify a reason
    re_match(".+", exclusions.rule_names[rule][workload][container])
}

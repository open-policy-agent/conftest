package main

exceptions = {"exception-name"}

func_name_msg(name) = ret {
	ret := sprintf("Resource Name '%s' contains dashes", [name])
}

deny_name[msg] {
	input.resource[_][name]
	contains(name, "-")
	msg := func_name_msg(name)
}

deny_resource_type[msg] {
	input.resource[type]
	type == "invalid_type"
	msg := sprintf("Resource Type '%s' is invalid", [type])
}

exclude_name[rules] {
	input.resource[_][name]
	exceptions[name]
	rules := [func_name_msg(name)]
}

exception[rules] {
	rules := ["resource_type"]
}

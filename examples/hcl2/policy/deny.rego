package main
import rego.v1

has_field(obj, field) if {
	obj[field]
}

deny contains msg if {
	some name
	some lb in input.resource.aws_alb_listener[name]
	lb.protocol == "HTTP"
	msg = sprintf("ALB `%v` is using HTTP rather than HTTPS", [name])
}

deny contains msg if {
	some name
	some rule in input.resource.aws_security_group_rule[name]
	rule.type == "ingress"
	contains(rule.cidr_blocks[_], "0.0.0.0/0")
	msg = sprintf("ASG `%v` defines a fully open ingress", [name])
}

deny contains msg if {
	some name
	some disk in input.resource.azurerm_managed_disk[name]
	has_field(disk, "encryption_settings")
	not disk.encryption_settings.enabled
	msg = sprintf("Azure disk `%v` is not encrypted", [name])
}

# Required tags for all AWS resources
required_tags := {"environment", "owner"}
missing_tags(resource) := {tag | tag := required_tags[_]; not resource.tags[tag]}

deny contains msg if {
	some aws_resource, name
	some resource in input.resource[aws_resource][name] # all resources
	startswith(aws_resource, "aws_") # only AWS resources
	missing := missing_tags(resource)
	count(missing) > 0

	msg = sprintf("AWS resource: %q named %q is missing required tags: %v", [aws_resource, name, missing])
}

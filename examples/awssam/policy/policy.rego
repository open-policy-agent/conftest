package main
import rego.v1

denylist := ["*"]

sensitive_denylist := [
	"password",
	"Password",
	"Pass",
	"pass",
]

runtime_denylist := [
	"python2.7",
	"node4.3",
]

check_resources(actions, denylist) if {
	endswith(actions[_], denylist[_])
}

check_sensitive(envs, denylist) if {
	contains(envs[_], denylist[_])
}

check_runtime(runtime, denylist) if {
	contains(runtime, denylist[_])
}

deny contains msg if {
	input.Resources.LambdaFunction.Properties.Runtime = "python2.7"
	msg = "python2.7 runtime not allowed"
}

deny contains msg if {
	input.Resources.LambdaFunction.Properties.Runtime = runtime
	check_runtime(runtime, runtime_denylist)
	msg = sprintf("%s runtime not allowed", [runtime])
}

deny contains msg if {
	input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Action = a
	check_resources(a, denylist)
	input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Effect = "Allow"
	msg = "excessive Action permissions not allowed"
}

deny contains msg if {
	input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Action = a
	is_string(a)
	endswith(a, "*")
	input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Effect = "Allow"
	msg = "excessive Action permissions not allowed"
}

deny contains msg if {
	input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Resource = a
	check_resources(a, denylist)
	input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Effect = "Allow"
	msg = "excessive Resource permissions not allowed"
}

deny contains msg if {
	input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Resource = a
	is_string(a)
	endswith(a, "*")
	input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Effect = "Allow"
	msg = "excessive Resource permissions not allowed"
}

deny contains msg if {
	input.Resources.LambdaFunction.Properties.Environment.Variables = a
	check_sensitive(a, sensitive_denylist)
	input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Effect = "Allow"
	msg = "Sensitive data not allowed in environment variables"
}

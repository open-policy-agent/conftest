package main

blacklist = [
    "*"
]

sensitive_blacklist = [
    "password",
    "Password",
    "Pass",
    "pass"
]

runtime_blacklist = [
    "python2.7",
    "node4.3"
]

check_resources(actions, blacklist) {
  endswith(actions[_], blacklist[_])
}

check_sensitive(envs, blacklist) {
  contains(envs[_], blacklist[_])
}

check_runtime(runtime, blacklist) {
  contains(runtime, blacklist[_])
}

deny[msg] {
  input.Resources.LambdaFunction.Properties.Runtime = "python2.7"
  msg = "python2.7 runtime not allowed"
}

deny[msg] {
  input.Resources.LambdaFunction.Properties.Runtime = runtime; check_runtime(runtime, runtime_blacklist)
  msg = sprintf("%s runtime not allowed", [runtime])
}

deny[msg] {
  input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Action = a; check_resources(a, blacklist)
  input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Effect = "Allow"
  msg = "excessive Action permissions not allowed"
}

deny[msg] {
  input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Action = a; is_string(a); endswith(a, "*")
  input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Effect = "Allow"
  msg = "excessive Action permissions not allowed"
}

deny[msg] {
  input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Resource = a; check_resources(a, blacklist)
  input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Effect = "Allow"
  msg = "excessive Resource permissions not allowed"
}

deny[msg] {
  input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Resource = a; is_string(a); endswith(a, "*")
  input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Effect = "Allow"
  msg = "excessive Resource permissions not allowed"
}

deny[msg] {
  input.Resources.LambdaFunction.Properties.Environment.Variables = a; check_sensitive(a, sensitive_blacklist)
  input.Resources.LambdaFunction.Properties.Policies[_].Statement[_].Effect = "Allow"
  msg = "Sensitive data not allowed in environment variables"
}
package main

denylist := ["openjdk"]

deny[msg] {
	some i
	input[i].Cmd == "from"
	val := input[i].Value
	contains(val[i], denylist[_])

	msg = sprintf("unallowed image found %s", [val])
}

package commands

denylist := [
	"apk",
	"apt",
	"pip",
	"curl",
	"wget",
]

deny[msg] {
	some i
	input[i].Cmd == "run"
	val := input[i].Value
	contains(val[_], denylist[_])

	msg := sprintf("unallowed commands found %s", [val])
}

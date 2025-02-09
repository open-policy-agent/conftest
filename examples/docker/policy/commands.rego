package commands
import rego.v1

denylist := [
	"apk",
	"apt",
	"pip",
	"curl",
	"wget",
]

deny contains msg if {
	some i
	input[i].Cmd == "run"
	val := input[i].Value
	contains(val[_], denylist[_])

	msg := sprintf("unallowed commands found %s", [val])
}

package main
import rego.v1

any_git_ignored if {
	entry := input[_]

	entry.Kind == "Path"
	entry.Value == ".git"
}

deny contains msg if {
	not any_git_ignored
	msg := ".git directories should be ignored"
}

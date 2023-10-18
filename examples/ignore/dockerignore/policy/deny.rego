package main

any_git_ignored {
	entry := input[_]

	entry.Kind == "Path"
	entry.Value == ".git"
}

deny[msg] {
	not any_git_ignored
	msg := ".git directories should be ignored"
}

package main

any_id_rsa_ignored {
	entry := input[i]

	entry.Kind == "Path"
	entry.Value == "id_rsa"
}

deny[msg] {
	not any_id_rsa_ignored
	msg = "id_rsa files should be ignored"
}

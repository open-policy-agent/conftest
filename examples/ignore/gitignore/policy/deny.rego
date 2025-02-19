package main
import rego.v1

any_id_rsa_ignored if {
	entry := input[_]

	entry.Kind == "Path"
	entry.Value == "id_rsa"
}

deny contains msg if {
	not any_id_rsa_ignored
	msg = "id_rsa files should be ignored"
}

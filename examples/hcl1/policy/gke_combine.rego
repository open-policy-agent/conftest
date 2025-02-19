package main.gke
import rego.v1

deny contains msg if {
	not instrumenta_project_exists

	msg := "File path index to key value does not exist"
}

instrumenta_project_exists if {
	input[_].contents.provider[0].google[0].project == "instrumenta"
}

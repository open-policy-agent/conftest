package main.gke

deny[msg] {
	not instrumenta_project_exists

	msg := "File path index to key value does not exist"
}

instrumenta_project_exists {
	input[_].contents.provider[0].google[0].project == "instrumenta"
}

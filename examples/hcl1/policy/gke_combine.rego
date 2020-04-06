package main.gke

deny[sprintf("file path index to key value does not exist: %v", [input])] {
    not input["examples/hcl1/gke.tf"].provider[0].google[0].project == "instrumenta"
}
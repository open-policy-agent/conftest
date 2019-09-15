package main

deny[sprintf("file path index to key value does not exist: %v", [input])] {
    not input["examples/terraform/gke.tf"].provider[0].google[0].project == "instrumenta"
}
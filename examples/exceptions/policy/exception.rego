package main
import rego.v1

exception contains rules if {
	input.kind = "Deployment"
	input.metadata.name = "can-run-as-root"

	rules = ["run_as_root"]
}

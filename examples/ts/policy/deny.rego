package main
import rego.v1

deny contains msg if {
	input.kind == "Pod"
	not input.metadata.labels.app
	msg = "Pods must provide an app label"
}

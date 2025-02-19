package main
import rego.v1

no_violations if {
	count(deny) == 0
}

test_no_missing_label if {
	deployment := {
		"kind": "Deployment",
		"metadata": {
			"name": "sample",
			"labels": {
				"app.kubernetes.io/name",
				"app.kubernetes.io/instance"
			}
		}
	}

	no_violations with input as deployment
}

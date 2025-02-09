package main
import rego.v1

no_violations if {
	count(deny) == 0
}

test_missing_required_label_fail if {
	deployment := {
		"kind": "Deployment",
		"metadata": {
			"name": "sample",
			"labels": {
				"app.kubernetes.io/instance"
			}
		}
	}

	no_violations with input as deployment
}

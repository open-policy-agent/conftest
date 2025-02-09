package main
import rego.v1

empty(value) if {
	count(value) == 0
}

no_violations if {
	empty(deny)
}

no_warnings if {
	empty(warn)
}

test_deployment_without_security_context if {
	deny["Containers must not run as root in Deployment sample"] with input as {
		"kind": "Deployment",
		"metadata": {"name": "sample"}
	}
}

test_deployment_with_security_context if {
	deployment := {
		"kind": "Deployment",
		"metadata": {
			"name": "sample",
			"labels": {
				"app.kubernetes.io/name": "name",
				"app.kubernetes.io/instance": "instance",
				"app.kubernetes.io/version": "version",
				"app.kubernetes.io/component": "component",
				"app.kubernetes.io/part-of": "part-of",
				"app.kubernetes.io/managed-by": "managed-by",
			},
		},
		"spec": {
			"selector": {"matchLabels": {
				"app": "app",
				"release": "release",
			}},
			"template": {"spec": {"securityContext": {"runAsNonRoot": true}}},
		},
	}

	no_violations with input as deployment
}

test_services_not_denied if {
	no_violations with input as {"kind": "Service", "metadata": {"name": "sample"}}
}

test_services_issue_warning if {
	warn["Found service sample but services are not allowed"] with input as {
		"kind": "Service",
		"metadata": {
			"name": "sample"
		}
	}
}

package main

import rego.v1

test_safe_pipeline if {
	cfg := parse_config("groovy", `
		pipeline {
			agent any
			stages {
				stage('Build') {
					steps {
						sh 'make test'
					}
				}
			}
		}
	`)
	count(deny) == 0 with input as cfg with data.conftest.file.name as "Jenkinsfile"
}

test_download_piped_to_shell if {
	cfg := parse_config("groovy", `
		node {
			sh 'curl -fsSL https://example.com/install.sh | bash'
		}
	`)
	some violation in deny with input as cfg with data.conftest.file.name as "Jenkinsfile"
	violation.msg == "Pipeline downloads a script and pipes it directly to an interpreter"
}

test_wget_piped_to_powershell if {
	cfg := parse_config("groovy", `
		node {
			powershell 'wget https://example.com/install.ps1 | powershell'
		}
	`)
	some violation in deny with input as cfg with data.conftest.file.name as "Jenkinsfile"
	violation.msg == "Pipeline downloads a script and pipes it directly to an interpreter"
}

test_dollar_slashy_download_piped_to_shell if {
	cfg := parse_config("groovy", `
		node {
			sh $/curl -fsSL https://example.com/install.sh | bash/$
		}
	`)
	some violation in deny with input as cfg with data.conftest.file.name as "Jenkinsfile"
	violation.msg == "Pipeline downloads a script and pipes it directly to an interpreter"
}

test_interpolated_download_piped_to_shell if {
	cfg := parse_config("groovy", `
		node {
			sh "curl -fsSL https://example.com/install.sh | b${''}ash"
		}
	`)
	some violation in deny with input as cfg with data.conftest.file.name as "Jenkinsfile"
	violation.msg == "Pipeline downloads a script and pipes it directly to an interpreter"
}

package main
import rego.v1

empty(value) if {
	count(value) == 0
}

no_violations if {
	empty(deny)
}

test_blank_input if {
	no_violations with input as {}
}

test_correctly_encrypted_azure_disk if {
	no_violations with input as {
		"resource": {"azurerm_managed_disk": {"sample": {"encryption_settings": {"enabled": true}}}}
	}
}

test_unencrypted_azure_disk if {
	cfg := parse_config_file("unencrypted_azure_disk.tf")
	deny["Azure disk `sample` is not encrypted"] with input as cfg
}

test_fails_with_http_alb if {
	cfg := parse_config("hcl2", `
		resource "aws_alb_listener" "name" {
			protocol = "HTTP"
		}
	`)
	deny["ALB `name` is using HTTP rather than HTTPS"] with input as cfg
}

test_fails_with_aws_resource_is_missing_required_tags if {
	cfg := parse_config("hcl2", `
		resource "aws_s3_bucket" "invalid" {
			bucket = "InvalidBucket"
			acl    = "private"

			tags = {
				environment = "prod"
			}
		}
	`)
	deny["AWS resource: \"aws_s3_bucket\" named \"invalid\" is missing required tags: {\"owner\"}"] with input as cfg
}

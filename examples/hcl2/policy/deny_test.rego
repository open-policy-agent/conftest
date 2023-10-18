package main

empty(value) {
	count(value) == 0
}

no_violations {
	empty(deny)
}

test_blank_input {
	no_violations with input as {}
}

test_correctly_encrypted_azure_disk {
	no_violations with input as {
		"resource": {"azurerm_managed_disk": {"sample": {"encryption_settings": {"enabled": true}}}}
	}
}

test_unencrypted_azure_disk {
	cfg := parse_config_file("unencrypted_azure_disk.tf")
	deny["Azure disk `sample` is not encrypted"] with input as cfg
}

test_fails_with_http_alb {
	cfg := parse_config("hcl2", `
		resource "aws_alb_listener" "name" {
			protocol = "HTTP"
		}
	`)
	deny["ALB `name` is using HTTP rather than HTTPS"] with input as cfg
}

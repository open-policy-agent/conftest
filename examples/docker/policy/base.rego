package main

deny[msg] {
  msg = sprintf("Terraform plan will change prohibited resources in the following namespaces: %v", [input])
}
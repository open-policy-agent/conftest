package main

denylist = [
  "google_iam",
  "google_container"
]

deny[msg] {
  check_resources(input.resource_changes, denylist)
  banned := concat(", ", denylist)
  msg = sprintf("Terraform plan will change prohibited resources in the following namespaces: %v", [banned])
}

# Checks whether the plan will cause resources with certain prefixes to change
check_resources(resources, disallowed_prefixes) {
  startswith(resources[_].type, disallowed_prefixes[_])
}

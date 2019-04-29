package main


blacklist = [
  "google_iam",
  "google_container"
]

deny[msg] {
  check_resources(input.resource_changes, blacklist)
  banned := concat(", ", blacklist)
  msg = sprintf("Terraform plan will change prohibited resources in the following namespaces: %v", [banned])
}

# Checks whether the plan will cause resources with certain prefixes to change
check_resources(resources, disallowed_prefixes) {
  startswith(resources[_].type, disallowed_prefixes[_])
}

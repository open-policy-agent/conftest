package main

exception[rules] {
  input.kind = "Deployment"
  input.metadata.name = "can-run-as-root"

  rules = ["run_as_root"]
}

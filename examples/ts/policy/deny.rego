package main

deny[msg] {
  input.kind == "Pod"
  not input.metadata.labels.app
  msg = "Pods must provide an app label"
}

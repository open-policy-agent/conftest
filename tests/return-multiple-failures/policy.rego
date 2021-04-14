package main

failures = ["one", "two", "three"]

deny[resource_name] {
  resource_name = failures[_]
}

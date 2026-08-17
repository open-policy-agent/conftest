package main
import rego.v1

# Check that file starts with default owner
deny contains msg if {
  input.entries[0].pattern != "*"
  msg := "first entry must assign default owner"
}

deny contains msg if {
  some entry in input.entries
  entry.pattern == "*.js"
  entry.owners != ["@my-org/js-devs"]
  msg := "JavaScript files must be owned by @my-org/js-devs"
}

deny contains msg if {
  some entry in input.entries
  entry.pattern == "*.go"
  entry.owners != ["@my-org/go-devs"]
  msg := "go files must be owned by @my-org/go-devs"
}

deny contains msg if {
  input.entries[-1].pattern != "very/important/file"
  msg := "last entry must assign dedicated owner for very/important/file"
}

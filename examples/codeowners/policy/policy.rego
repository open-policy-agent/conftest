package main
import rego.v1

# Check that file starts with default owner
deny contains msg if {
  input.rules[0].pattern != "*"
  msg := "first rule must assign default owner"
}

deny contains msg if {
  some rule in input.rules
  rule.pattern == "*.js"
  rule.owners != [{
    "type": "team",
    "value": "my-org/js-devs",
  }]
  msg := "JavaScript files must be owned by my-org/js-devs"
}

deny contains msg if {
  some rule in input.rules
  rule.pattern == "*.go"
  rule.owners != [{
    "type": "team",
    "value": "my-org/go-devs",
  }]
  msg := "go files must be owned by my-org/go-devs"
}

deny contains msg if {
  input.rules[-1].pattern != "very/important/file"
  msg := "last rule must assign dedicated owner for very/important/file"
}

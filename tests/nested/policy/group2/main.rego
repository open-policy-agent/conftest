package group2
import rego.v1

deny contains msg if {
	input.hello = "world"
	msg = "nested json group2 failed"
}

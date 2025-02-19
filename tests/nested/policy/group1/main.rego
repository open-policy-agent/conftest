package group1
import rego.v1

deny contains msg if {
	input.hello = "world"
	msg = "nested json group1 failed"
}

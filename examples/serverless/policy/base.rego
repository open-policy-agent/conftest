package main
import rego.v1

deny contains msg if {
	input.provider.runtime = "python2.7"
	msg = "Python 2.7 cannot be the default provider runtime"
}

runtime contains msg if {
	input.functions[_].runtime = name
}

deny contains msg if {
	runtime["python2.7"]
	msg = "Python 2.7 cannot be used as the runtime for functions"
}

deny contains msg if {
	not has_field(input.provider.tags, "author")
	msg = "Should set provider tags for author"
}

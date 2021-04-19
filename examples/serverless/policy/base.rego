package main

deny[msg] {
	input.provider.runtime = "python2.7"
	msg = "Python 2.7 cannot be the default provider runtime"
}

runtime[name] {
	input.functions[i].runtime = name
}

deny[msg] {
	runtime["python2.7"]
	msg = "Python 2.7 cannot be used as the runtime for functions"
}

deny[msg] {
	not has_field(input.provider.tags, "author")
	msg = "Should set provider tags for author"
}

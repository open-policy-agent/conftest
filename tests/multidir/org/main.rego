package main

deny[msg] {
	input.foo = "bar"
	msg = "Org policy forbids foo=bar"
}

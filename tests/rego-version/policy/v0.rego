package main

deny[msg] {
    input.bar == "baz"
    msg := "foo"
}

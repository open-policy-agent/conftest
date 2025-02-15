package main

deny contains msg if {
    input.bar == "baz"
    msg := "foo"
}

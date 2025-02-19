package main
import rego.v1

allow if {
    input.b == "foo"
    a := 1
    b := 2
    x := {
        "a": a,
        "b": "bar",
    }
    c := 3
}

validate(x, y) if {
	input.test == x
} else := false if {
    input.test == "foo"
    allow
}

test(x, y, z) if {
	input.test == x
} else if {
	input.test == y
} else if {
	input.test == z
}

deny contains msg if {
    test("foo", "bar", "baz")
    validate("foo", "bar")
    msg = "deployment objects should have validated"
}

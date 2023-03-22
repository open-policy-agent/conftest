package main

allow {
    input.b == "foo"
    a := 1
    b := 2
    x := {
        "a": a,
        "b": "bar",
    }
    c := 3
}

validate(x, y) {
	input.test == x
} else := false {
    input.test == "foo"
    allow
}

test(x, y, z) {
	input.test == x
} else {
	input.test == y
} else {
	input.test == z
}

deny[msg] {
    test("foo", "bar", "baz")
    validate("foo", "bar")
    msg = "deployment objects should have validated"
}
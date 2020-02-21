package lib.common

sample_list := [
	"marshall",
	"mathers",
	"haley",
	"kim",
]

test_list_contains_value_pass {
	list_contains_value(sample_list, "mathers")
}

test_list_contains_value_false {
	list_contains_value(sample_list, "biggie") == false
}

test_list_contains_value_not {
	not list_contains_value(sample_list, "biggie")
}

sample_object_for_has_key := {
	"luke": "skywalker",
	"obiwan": "kenobi",
}

# Pass example
test_has_field_pass {
	has_field(sample_object_for_has_key, "luke")
}

# False example
test_has_field_false {
	false == has_field(sample_object_for_has_key, "kylo")
}

# Not example
test_has_field_not {
	not has_field(sample_object_for_has_key, "kylo")
}

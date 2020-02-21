package lib.common

list_contains_value(list, item) {
	list_item = list[_]
	list_item == item
}

else = false {
	true # It will return false unconditionally, if previous function is not true
}

has_field(obj, field) {
	obj[field]
}

else = false {
	true # It will return false unconditionally, if previous function is not true
}

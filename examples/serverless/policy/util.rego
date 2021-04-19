package main

# has_field returns whether an object has a field
has_field(object, field) {
	object[field]
}

# False is a tricky special case, as false responses would create an undefined document unless
# they are explicitly tested for
has_field(object, field) {
	object[field] == false
}

has_field(object, field) = false {
	not object[field]
	not object[field] == false
}

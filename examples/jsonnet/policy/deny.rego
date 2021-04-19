package main

deny[msg] {
	not input.concat_array < 3
	msg = "Concat array should be less than 3"
}

deny[msg] {
	not input.obj_member = true
	msg = "Object member should be true"
}

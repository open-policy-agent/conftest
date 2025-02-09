package main
import rego.v1

deny contains msg if {
	not input.concat_array < 3
	msg = "Concat array should be less than 3"
}

deny contains msg if {
	not input.obj_member = true
	msg = "Object member should be true"
}

package main
import rego.v1

deny contains {"msg": msg} if {
    input.test_field == 123
    msg := "some error"
}

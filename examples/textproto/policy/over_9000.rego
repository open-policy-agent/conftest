package main
import rego.v1

deny contains {"msg": msg} if {
    input.number <= 9000
    msg := sprintf("%s: Power level must be over 9000", [input.name])
}

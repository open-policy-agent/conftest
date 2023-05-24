package main

deny[{"msg": msg}] {
    input.number <= 9000
    msg := sprintf("%s: Power level must be over 9000", [input.name])
}

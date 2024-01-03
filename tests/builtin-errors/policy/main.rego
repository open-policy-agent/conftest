package main

deny[{"msg": msg}] {
    input.test_field == 123
    msg := "some error"
}

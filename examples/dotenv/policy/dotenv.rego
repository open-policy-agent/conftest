package main

deny[msg] {
    input.APP_NAME == ""
    msg = "APP_NAME must be set"
}

deny[msg] {
    input.MYSQL_USER == "root"
    msg = "MYSQL_USER should not be root"
}

package main
import rego.v1

deny contains msg if {
    input.APP_NAME == ""
    msg = "APP_NAME must be set"
}

deny contains msg if {
    input.MYSQL_USER == "root"
    msg = "MYSQL_USER should not be root"
}

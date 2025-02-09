package main
import rego.v1

empty(value) if {
	count(value) == 0
}

no_violations if {
	empty(deny)
}

test_app_name_is_not_set if {
  deny["APP_NAME must be set"] with input as { "APP_NAME": "" }
}

test_app_name_is_set if {
  no_violations with input as { "APP_NAME": "test" }
}

test_mysql_user_is_root if {
  deny["MYSQL_USER should not be root"] with input as { "MYSQL_USER": "root" }
}

test_mysql_user_is_not_root if {
  no_violations with input as { "MYSQL_USER": "user1" }
}

package main

empty(value) {
	count(value) == 0
}

no_violations {
	empty(deny)
}

test_app_name_is_not_set {
  deny["APP_NAME must be set"] with input as { "APP_NAME": "" }
}

test_app_name_is_set {
  no_violations with input as { "APP_NAME": "test" }
}

test_mysql_user_is_root {
  deny["MYSQL_USER should not be root"] with input as { "MYSQL_USER": "root" }
}

test_mysql_user_is_not_root {
  no_violations with input as { "MYSQL_USER": "user1" }
}

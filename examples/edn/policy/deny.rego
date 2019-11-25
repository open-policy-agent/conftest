package main

deny[msg] {
  input[":env"] = ":development"
  input[":log"] != ":debug"
  msg = "Applications in the development environment should have debug logging"
}

deny[msg] {
  input[":env"] = ":production"
  input[":log"] != ":error"
  msg = "Applications in the production environment should have error only logging"
}
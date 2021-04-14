package main

deny[msg] {
  input.baz = "qux"
  msg = "Team policy forbids baz=qux"
}

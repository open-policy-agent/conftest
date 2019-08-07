package main

deny[msg] {
  not sprintf("%s", input) != "null" 
  msg = sprintf("Deployment %s", input)
}

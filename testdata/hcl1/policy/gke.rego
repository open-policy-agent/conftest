package main

deny[sprintf("could not find any resources in: %v", [input])] {
  count(input.resource) == 0
}
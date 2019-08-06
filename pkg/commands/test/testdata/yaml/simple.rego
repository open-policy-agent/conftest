package main

deny_when_simple_is_not_true["Value 'simple' wasn't true"] {
  input.simple != true
}
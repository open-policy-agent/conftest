package main

deny_wrongname[msg] {
  input[_].metadata.name == "hello-kubernetes"
  msg = sprintf("nothing to see here %v", [input]) 
}

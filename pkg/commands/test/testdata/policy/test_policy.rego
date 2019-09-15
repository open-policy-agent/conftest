package main

deny_wrongname[msg] {
  input.metadata.name == "hello-kubernetes"
  msg = sprintf("nothing to see here %v", [input]) 
}

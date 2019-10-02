package main

blacklist = [
  "openjdk"
]

deny[msg] {
  input[i].Cmd == "from"
  val := input[i].Value
  contains(val[i], blacklist[_])

  msg = sprintf("blacklisted image found %s", [val])
}

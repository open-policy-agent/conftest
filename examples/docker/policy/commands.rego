package commands

blacklist = [
  "apk",
  "apt",
  "pip",
  "curl",
  "wget",
]

deny[msg] {
  input[i].Cmd == "run"
  val := input[i].Value
  contains(val[_], blacklist[_])

  msg = sprintf("blacklisted commands found %s", [val])
}

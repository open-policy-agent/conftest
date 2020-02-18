package main


version {
  to_number(input.version)
}


deny[msg] {
  endswith(input.services[_].image, ":latest")
  msg = "No images tagged latest"
}

deny[msg] {
  version < 3.5
  msg = "Must be using at least version 3.5 of the Compose file format"
}

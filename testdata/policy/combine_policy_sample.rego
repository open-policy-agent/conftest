package combine

deny_when_doesnt_have_multiple_files[sprintf("unable to find a gke container resource", [input])] {
  not check_for_multifiles
  not check_for_resources
}

check_for_multifiles {
  numberOfFiles = 2
  not input[numberOfFiles]
  input[numberOfFiles - 1]
}

check_for_resources {
  keys = [
    key |
    input[_].resource[keyName] == true
    startswith(keyName, "google_container")
    key := keyName
  ]
  count(keys) == 2
}

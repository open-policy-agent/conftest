package combine

deny_when_doesnt_have_multiple_files[sprintf("unable to find a gke container resource", [input])] {
  not check_for_multifiles
}

check_for_multifiles {
  numberOfFiles = 2
  not input[numberOfFiles]
  input[numberOfFiles - 1]
}

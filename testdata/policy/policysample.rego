package main


deny[sprintf("unable to find a gke container resource %s %s", [input.Resources[i].Type, input.Resources[i].Name])] {
  not check_for_container_resources
}

check_for_container_resources {
    containers := [
      container |
      some i
      name := input.Resources[i].Name
      type := input.Resources[i].Type
      startswith(type, "google_container")
      container := type
    ]
    count(containers) == 2
}

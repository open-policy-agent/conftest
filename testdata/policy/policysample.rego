package main

deny_when_not_correct_containers[sprintf("unable to find a gke container resource", [input])] {
  not check_for_container_resources
}

check_for_container_resources {
    count(input["resource"]) == 2
}

package main

warn[msg] {
  not combinedObject
  msg = "Found name 'service' and weather 'bad' in combined object"
}

warn[msg] {
  not namespaceCollision
  msg = "properly handles namespaces collisions"
}

combinedObject {
  some i
  input[i].name == "service"
  some j
  input[j].weather == "bad"
}

namespaceCollision {
  input[_].name == "service"
  input[_].name == "notService"
}
package main

test_service_denied {
  input := {
    "kind": "Service",
    "metadata": {
      "name": "sample"
      }, 
      "spec": {
        "type": "LoadBalancer",
        "ports": [{ "port":  22 }]
      }
  }

  deny["Cannot expose port 22 on LoadBalancer. Denied ports: [22, 21]"] with input as input
}

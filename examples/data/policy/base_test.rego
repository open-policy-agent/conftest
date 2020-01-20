package main

test_services_not_denied {
  deny["Cannot expose one of the following ports on a LoadBalancer [22]"] with input as {"kind": "Service", "metadata": { "name": "sample" }, "spec": { "type": "LoadBalancer", "ports": [{ "port":  22 }]}}
}

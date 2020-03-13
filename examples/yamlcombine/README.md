# Kubernetes multi-document yaml example

This example demonstrates using the `--combine` flag to test for issues arrising from rule violations which requires checking multiple documents. In this case, the policy enforces that any [Ingress](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#ingress-v1beta1-networking-k8s-io) which has a rule pointing at a [Service](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#service-v1-core) should use one of the named ports defined in the service.

To run this example, simply invoke it with the following command:

```
conftest test --combine ingresses.yaml services.yaml
```

The expected output is:

```
FAIL - Combined - Ingress 'my-service-ingress' (in ingresses.yaml) points at port 'https' in service 'my-service' (in services.yaml). However this service doesn't define this port (available ports: http)
FAIL - Combined - Ingress 'my-service-ingress' (in ingresses.yaml) points at port 'http' in service 'my-other-service' (in services.yaml). However this service doesn't define this port (available ports: https)
```

## Warning
The current behaviour of `conftest` with the `--combine` has an issue related to multi-document yaml files. Specifically, should some of the input files only contain a single document (which can easily happen with the output of simple helm charts for example), the input for this file will suddenly *not* be an array of sub-documents and this policy will not work for that file. See #106 for more information. You can see this problem in acion by removing the second ingress from `ingress.yaml`, even though that ingress yields no errors, removing it will seemingly fix all failures.

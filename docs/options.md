# Options

## `--input`

`conftest` normally detects input type with the file extension, but you can force to use a different one with the `--input` flag (`-i`).

For the available parsers, take a look at: [parsers](parser). For instance:

```console
$ conftest test -p examples/hcl2/policy examples/hcl2/terraform.tf -i hcl2
FAIL - examples/hcl2/terraform.tf - ALB `my-alb-listener` is using HTTP rather than HTTPS
FAIL - examples/hcl2/terraform.tf - ASG `my-rule` defines a fully open ingress
FAIL - examples/hcl2/terraform.tf - Azure disk `source` is not encrypte
```

The `--input` flag can also be a good way to see how different input types would be parsed:

```console
conftest parse examples/hcl2/terraform.tf -i hcl2
```

### Multi-input type

`conftest` supports multiple different input types in a single call.

```console
$ conftest test examples/multitype/grafana.ini examples/multitype/kubernetes.yaml -p examples/multitype
```

## `--combine`

This flag introduces *BREAKING CHANGES* in how `conftest` provides input to rego policies. However, you may find it useful to as you can now compare multiple values from different configurations simultaneously.

The `--combine` flag combines files into one `input` data structure. The structure is a `map` where each indice is the file path of the file being evaluated.

Let's try it!

Save the following as `policy/combine.rego`:

```rego
package main


deny[msg] {
  deployment := input["deployment.yaml"]["spec"]["selector"]["matchLabels"]["app"]
  service := input["service.yaml"]["spec"]["selector"]["app"]

  deployment != service

  msg = sprintf("Expected these values to be the same but received %v for deployment and %v for service", [deployment, service])
}
```

Save the following as `service.yaml`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: hello-kubernetes
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: goodbye-kubernetes
```

and lastly, save the following as `deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-kubernetes
spec:
  replicas: 3
  selector:
    matchLabels:
      app: hello-kubernetes
```

With the above files created, run the following command:

```console
$ conftest test --combine deployment.yaml service.yaml

FAIL - Combined-configs (multi-file) - Expected these values to be the same but received hello-kubernetes for deployment and goodbye-kubernetes for service
```

This is just the tip of the iceberg. Now you can ensure that duplicate values match across the entirety of your configuration files.

## `--data`

Sometimes policies require additional data in order to determine an answer. For example, a whitelist of allowed resources that can be created. Instead of hardcoding this information inside of your policy, conftest allows passing paths to data files with the `--data` flag, e.g.

```console
conftest test -p examples/data/policy -d examples/data/exclusions examples/data/service.yaml
```

The paths at the flag are recursively searched for JSON and YAML files. Data can be imported as follows:

Given the following yaml file:

```yaml
services:
  ports:
  - 22
```

This can be imported into your policy:

```rego
import data.services

ports := services.ports
```

## `--output`

The output of `conftest` can be configured using the `--output` flag (`-o`).

As of today `conftest` supports the following output types:

- Plaintext `--output=stdout`
- JSON: `--output=json`
- [TAP](https://testanything.org/): `--output=tap`
- Table `--output=table`

### Plaintext

```console
$ conftest test -p examples/kubernetes/policy examples/kubernetes/deployment.yaml
FAIL - examples/kubernetes/deployment.yaml - Containers must not run as root in Deployment hello-kubernetes
FAIL - examples/kubernetes/deployment.yaml - Deployment hello-kubernetes must provide app/release labels for pod selectors
FAIL - examples/kubernetes/deployment.yaml - hello-kubernetes must include Kubernetes recommended labels: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels
```

### JSON

```console
$ conftest test -o json -p examples/kubernetes/policy examples/kubernetes/deployment.yaml
[
        {
                "filename": "examples/kubernetes/deployment.yaml",
                "warnings": [],
                "failures": [
                        "hello-kubernetes must include Kubernetes recommended labels: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels ",
                        "Containers must not run as root in Deployment hello-kubernetes",
                        "Deployment hello-kubernetes must provide app/release labels for pod selectors"
                ]
        }
]
```

### TAP

```console
$ conftest test -o tap -p examples/kubernetes/policy examples/kubernetes/deployment.yaml
1..3
not ok 1 - examples/kubernetes/deployment.yaml - Containers must not run as root in Deployment hello-kubernetes
not ok 2 - examples/kubernetes/deployment.yaml - Deployment hello-kubernetes must provide app/release labels for pod selectors
not ok 3 - examples/kubernetes/deployment.yaml - hello-kubernetes must include Kubernetes recommended labels: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels
```

### Table

```console
$ conftest test -p examples/kubernetes/policy examples/kubernetes/service.yaml -o table
+---------+----------------------------------+--------------------------------+
| RESULT  |               FILE               |            MESSAGE             |
+---------+----------------------------------+--------------------------------+
| success | examples/kubernetes/service.yaml |                                |
| warning | examples/kubernetes/service.yaml | Found service hello-kubernetes |
|         |                                  | but services are not allowed   |
+---------+----------------------------------+--------------------------------+
```

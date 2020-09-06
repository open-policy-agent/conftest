# Options

## Configuration

Options for `conftest` can be configured on the command line as flag arguments, environment variables, or a configuration file.

When multiple configuration sources are present, `conftest` will process the configuration values in the following order:

1. Flag Arguments
1. Environment Variables
1. Configuration File

When using environment variables, the environment variable should be the same name as the flag, prefixed with `CONFTEST_`. For example, to set the policy directory, the environment variable would be `CONFTEST_POLICY`.

When using a configuration file, the configuration file should be in the working directory for `conftest` and named `conftest.toml`. An example can be found below:

```toml
# You can override the directory in which to store and look for policies
policy = "tests"

# You can override the namespace which to search for rules
namespace = "conftest"
```

## `--policy`

`conftest` will by default look for policies in the subdirectory `policy`, but you can point out another directory with the `--policy` (or `-p`) flag. This flag can also be repeated, if you have multiple directories with policy files, for example one with organization-wide policies, and another one with team-specific policies. Keep in mind when using multiple directories that the rules are treated as if they'd all been read from the same directory, and make sure any rules sharing names are compatible.

```console
$ conftest test files/ # Reads rules from policy/ directory (default)
$ conftest test -p my-policies files/ # Reads rules from my-policies/
$ conftest test -p my-policies -p org-policies files/ # Reads rules from my-policies/ *and* org-policies/

## `--input`

`conftest` normally detects input type with the file extension, but you can force to use a different one with the `--input` flag (`-i`).

For the available parsers, take a look at: [parsers](parser). For instance:

```console
$ conftest test -p examples/hcl/policy examples/unknown_extension/unknown.xyz -i hcl2
FAIL - examples/unknown_extension/unknown.xyz - ALB `my-alb-listener` is using HTTP rather than HTTPS
FAIL - examples/unknown_extension/unknown.xyz - ASG `my-rule` defines a fully open ingress
FAIL - examples/unknown_extension/unknown.xyz - Azure disk `source` is not encrypted
```

The `--input` flag can also be combined with `parse` command:

```console
# The outputs will be different since we harvest the input with different parsers
conftest parse examples/unknown_extension/unknown.xyz -i hcl2
conftest parse examples/unknown_extension/unknown.xyz -i hcl1
```

### Multi-input type

`conftest` supports multiple different input types in a single call.

```console
$ conftest test examples/multitype/grafana.ini examples/multitype/deployment.yaml -p examples/multitype
```

## `--ignore`

When the given input file lists/directories contain unwanted file extensions/folders, the `--ignore` flag helps you to filter them.
A regexp pattern can be passed as the flag value for filtering. Here's some examples:

```console
conftest test -p examples/test/ test/ --ignore="parent/"
conftest test -p examples/test/ test/ --ignore="child/"
conftest test -p examples/test/ test/ --ignore=".*.yaml"
conftest test -p examples/test/ test/ --ignore=".*.cue|.*.yaml"
```

## `--combine`

This flag introduces *BREAKING CHANGES* in how `conftest` provides input to rego policies. However, you may find it useful to use as it allows you to compare multiple values from different configurations simultaneously.

The `--combine` flag combines files into one `input` data structure. The structure is a `map` where each index is the file path of the file being evaluated.

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
FAIL - examples/kubernetes/deployment.yaml - Found deployment hello-kubernetes but deployments are not allowed
```

### JSON

```console
$ conftest test -o json -p examples/kubernetes/policy examples/kubernetes/deployment.yaml
[
        {
                "filename": "examples/kubernetes/deployment.yaml",
                "warnings": [],
                "failures": [
                        {
                                "msg": "Containers must not run as root in Deployment hello-kubernetes"
                        },
                        {
                                "msg": "Deployment hello-kubernetes must provide app/release labels for pod selectors"
                        },
                        {
                                "msg": "hello-kubernetes must include Kubernetes recommended labels: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels"
                        },
                        {
                                "msg": "Found deployment hello-kubernetes but deployments are not allowed",
                                "metadata": {
                                        "details": {}
                                }
                        }
                ],
                "successes": [
                        {
                                "msg": "data.main.warn"
                        }
                ]
        }
]
```

### TAP

```console
$ conftest test -o tap -p examples/kubernetes/policy examples/kubernetes/deployment.yaml
1..5
not ok 1 - examples/kubernetes/deployment.yaml - Containers must not run as root in Deployment hello-kubernetes
not ok 2 - examples/kubernetes/deployment.yaml - Deployment hello-kubernetes must provide app/release labels for pod selectors
not ok 3 - examples/kubernetes/deployment.yaml - hello-kubernetes must include Kubernetes recommended labels: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels
not ok 4 - examples/kubernetes/deployment.yaml - Found deployment hello-kubernetes but deployments are not allowed
# Successes
ok 5 - examples/kubernetes/deployment.yaml - data.main.warn
```

### Table

```console
$ conftest test -p examples/kubernetes/policy examples/kubernetes/service.yaml -o table
+---------+----------------------------------+--------------------------------+
| RESULT  |               FILE               |            MESSAGE             |
+---------+----------------------------------+--------------------------------+
| success | examples/kubernetes/service.yaml | data.main.deny                 |
| success | examples/kubernetes/service.yaml | data.main.violation            |
| success | examples/kubernetes/service.yaml |                                |
| success | examples/kubernetes/service.yaml |                                |
| warning | examples/kubernetes/service.yaml | Found service hello-kubernetes |
|         |                                  | but services are not allowed   |
+---------+----------------------------------+--------------------------------+
```

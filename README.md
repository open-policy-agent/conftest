# Conftest

[![CircleCI](https://circleci.com/gh/instrumenta/conftest.svg?style=svg)](https://circleci.com/gh/instrumenta/conftest)

- Join us on [the Open Policy Agent Slack](https://slack.openpolicyagent.org/) in the #conftest channel

## What

`conftest` is a utility to help you write tests against structured configuration data. For instance you could
write tests for your Kubernetes configurations, or Tekton pipeline definitions, Terraform code, Serverless configs
 or any other structured data.

`conftest` relies on the Rego language from [Open Policy Agent](https://www.openpolicyagent.org/) for writing
the assertions. You can read more about Rego in [How do I write policies](https://www.openpolicyagent.org/docs/how-do-i-write-policies.html)
in the Open Policy Agent documentation.

## Usage

`conftest` allows you to write policies using Open Policy Agent/rego and apply them to one or
more configuration files.

As of today `conftest` supports:

* YAML
* JSON
* INI
* TOML
* HOCON
* HCL
* CUE
* Dockerfile
* HCL2 (Experimental)
* EDN
* VCL
* XML

Policies by default should be placed in a directory
called `policy` but this can be overridden.

For instance, save the following as `policy/deployment.rego`:

```rego
package main

deny[msg] {
  input.kind = "Deployment"
  not input.spec.template.spec.securityContext.runAsNonRoot = true
  msg = "Containers must not run as root"
}

deny[msg] {
  input.kind = "Deployment"
  not input.spec.selector.matchLabels.app
  msg = "Containers must provide app label for pod selectors"
}
```

By default `conftest` looks for `deny` and `warn` rules in the `main` namespace. This can be
altered by running `--namespace` or provided on the configuration file.

Assuming you have a Kubernetes deployment in `deployment.yaml` you can run `conftest` like so:

```console
$ conftest test deployment.yaml
FAIL - deployment.yaml - Containers must not run as root
FAIL - deployment.yaml - Deployments are not allowed
```

`conftest` can also be used with stdin:

```console
$ cat deployment.yaml | conftest test -
FAIL - Containers must not run as root
FAIL - Deployments are not allowed
```

Note that `conftest` isn't specific to Kubernetes. It will happily let you write tests for any
configuration files.

### --input flag

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

#### Multi input type

`conftest` supports multiple different input types in a single call.

```console
$ conftest test examples/multitype/grafana.ini examples/multitype/kubernetes.yaml -p examples/multitype
```

### --combine flag

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

### --data flag

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

### --output flag

The output of `conftest` can be configured using the `--output` flag (`-o`).

As of today `conftest` supports the following output types:

- Plaintext `--output=stdout`
- JSON: `--output=json`
- [TAP](https://testanything.org/): `--output=tap`
- Table `--output=table`

#### Example Output

##### Plaintext

```console
$ conftest test -p examples/kubernetes/policy examples/kubernetes/deployment.yaml
FAIL - examples/kubernetes/deployment.yaml - Containers must not run as root in Deployment hello-kubernetes
FAIL - examples/kubernetes/deployment.yaml - Deployment hello-kubernetes must provide app/release labels for pod selectors
FAIL - examples/kubernetes/deployment.yaml - hello-kubernetes must include Kubernetes recommended labels: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels
```

##### JSON

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

##### TAP

```console
$ conftest test -o tap -p examples/kubernetes/policy examples/kubernetes/deployment.yaml
1..3
not ok 1 - examples/kubernetes/deployment.yaml - Containers must not run as root in Deployment hello-kubernetes
not ok 2 - examples/kubernetes/deployment.yaml - Deployment hello-kubernetes must provide app/release labels for pod selectors
not ok 3 - examples/kubernetes/deployment.yaml - hello-kubernetes must include Kubernetes recommended labels: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels
```

##### TABLE

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

## Examples

You can find examples using various other tools in the `examples` directory, including:

* [AWS SAM Framework](examples/awssam)
* [CUE](examples/cue)
* [Docker compose](examples/compose)
* [Dockerfile](examples/docker)
* [EDN](examples/edn)
* [HCL2](examples/hcl2)
* [HOCON](examples/hocon)
* [INI](examples/ini)
* [GitLab](examples/ci)
* [Kubernetes](examples/kubernetes)
* [Kustomize](examples/kustomize)
* [Multitype](examples/multitype)
* [Serverless Framework](examples/serverless)
* [Tekton](examples/tekton)
* [Terraform](examples/terraform)
* [Traefik](examples/traefik)
* [Typescript](examples/ts)
* [VCL](examples/vcl)
* [XML](examples/xml)

## Configuration and external policies

Policies are often reusable between different projects, and `conftest` supports a mechanism
to specify dependent policies as well as download them. The format reuses the [Bundle defined
by Open Policy Agent](https://www.openpolicyagent.org/docs/latest/bundles).

You can download individual policies directly:

```console
conftest pull instrumenta.azurecr.io/test
```

Pull also supports other policy locations, such as git or https. Under the hood conftest leverages [go-getter](https://github.com/hashicorp/go-getter) to download policies. For example, to download a policy via https:

```console
conftest pull https://raw.githubusercontent.com/instrumenta/conftest/master/examples/compose/policy/deny.rego
```

Policies can be stored in OCI-compatible registries. You can read more about this idea in
[this post](https://stevelasker.blog/2019/01/25/cloud-native-artifact-stores-evolve-from-container-registries/).
Conftest supports storing policies using this mechanism leveraging [ORAS](https://github.com/deislabs/oras).

If you have a compatible OCI registry you can also push new policy bundles like so:

```console
conftest push instrumenta.azurecr.io/test
conftest push 127.0.0.1:5000/test
conftest push <some-other-supported-registry>/test
```

OCI bundles can be pulled as well:

```console
conftest pull instrumenta.azurecr.io/test
conftest pull 127.0.0.1:5000/test
conftest pull oci://<some-other-supported-registry>/test
```

The Azure registy and 127.0.0.1:5000 (The local [Docker Registry](https://github.com/docker/distribution)) are special cases where the URL does not need to be prefixed with the scheme `oci://`, in all other cases the scheme needs to be provided in the URL.

If you want to download the latest policies and run the tests in one go, you can do so with:

```console
conftest test --update <url(s)> <file-to-test>
```

`conftest` also supports a simple configuration file which can be used to store
configuration settings for the `conftest` command.

Create a `conftest.toml` configuration file like the following:

```toml
# You can override the directory in which to store and look for policies
policy = "tests"
# You can override the namespace which to search for rules
namespace = "conftest"
```

## Debugging queries

When working on more complex queries (or when learning rego), it's useful to see exactly how the policy is
applied. For this purpose you can use the `--trace` flag. This will output a large trace from Open Policy Agent
like the following:

<details>
<summary>Example of trace</summary>

```console
$ conftest test --trace deployment.yaml
FAIL - deployment.yaml - Deployment hello-kubernetes must provide app/release labels for pod selectors
TRAC - deployment.yaml - Enter data.main.deny = _
TRAC - deployment.yaml - | Eval data.main.deny = _
TRAC - deployment.yaml - | Index data.main.deny = _ matched 3 rules)
TRAC - deployment.yaml - | Enter data.main.deny
TRAC - deployment.yaml - | | Eval data.kubernetes.is_deployment
TRAC - deployment.yaml - | | Index data.kubernetes.is_deployment (matched 1 rule)
TRAC - deployment.yaml - | | Enter data.kubernetes.is_deployment
TRAC - deployment.yaml - | | | Eval input.kind = "Deployment"
TRAC - deployment.yaml - | | | Exit data.kubernetes.is_deployment
TRAC - deployment.yaml - | | Eval not data.main.labels
TRAC - deployment.yaml - | | Enter data.main.labels
TRAC - deployment.yaml - | | | Eval data.main.labels
TRAC - deployment.yaml - | | | Index data.main.labels matched 2 rules)
TRAC - deployment.yaml - | | | Enter data.main.labels
TRAC - deployment.yaml - | | | | Eval input.metadata.labels["app.kubernetes.io/name"]
TRAC - deployment.yaml - | | | | Eval input.metadata.labels["app.kubernetes.io/instance"]
TRAC - deployment.yaml - | | | | Fail input.metadata.labels["app.kubernetes.io/instance"]
TRAC - deployment.yaml - | | | | Redo input.metadata.labels["app.kubernetes.io/name"]
TRAC - deployment.yaml - | | | Enter data.main.labels
TRAC - deployment.yaml - | | | | Eval input.spec.selector.matchLabels.app
TRAC - deployment.yaml - | | | | Eval input.spec.selector.matchLabels.release
TRAC - deployment.yaml - | | | | Fail input.spec.selector.matchLabels.release
TRAC - deployment.yaml - | | | | Redo input.spec.selector.matchLabels.app
TRAC - deployment.yaml - | | | Fail data.main.labels
TRAC - deployment.yaml - | | Eval __local9__ = data.main.name
TRAC - deployment.yaml - | | Index __local9__ = data.main.name matched 3 rules)
TRAC - deployment.yaml - | | Enter data.main.name
TRAC - deployment.yaml - | | | Eval true
TRAC - deployment.yaml - | | | Eval __local5__ = input.metadata.name
TRAC - deployment.yaml - | | | Exit data.main.name
TRAC - deployment.yaml - | | Eval sprintf("%s must include Kubernetes recommended labels: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels ", [__local9__], __local2__)
TRAC - deployment.yaml - | | Eval msg = __local2__
TRAC - deployment.yaml - | | Exit data.main.deny
TRAC - deployment.yaml - | Redo data.main.deny
TRAC - deployment.yaml - | | Redo msg = __local2__
TRAC - deployment.yaml - | | Redo sprintf("%s must include Kubernetes recommended labels: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels ", [__local9__], __local2__)
TRAC - deployment.yaml - | | Redo __local9__ = data.main.name
TRAC - deployment.yaml - | | Redo data.main.name
TRAC - deployment.yaml - | | | Redo __local5__ = input.metadata.name
TRAC - deployment.yaml - | | | Redo true
TRAC - deployment.yaml - | | Enter data.main.name
TRAC - deployment.yaml - | | | Eval true
TRAC - deployment.yaml - | | | Eval __local6__ = input.metadata.name
TRAC - deployment.yaml - | | | Exit data.main.name
TRAC - deployment.yaml - | | Redo data.main.name
TRAC - deployment.yaml - | | | Redo __local6__ = input.metadata.name
TRAC - deployment.yaml - | | | Redo true
TRAC - deployment.yaml - | | Enter data.main.name
TRAC - deployment.yaml - | | | Eval true
TRAC - deployment.yaml - | | | Eval __local4__ = input.metadata.name
TRAC - deployment.yaml - | | | Exit data.main.name
TRAC - deployment.yaml - | | Redo data.main.name
TRAC - deployment.yaml - | | | Redo __local4__ = input.metadata.name
TRAC - deployment.yaml - | | | Redo true
TRAC - deployment.yaml - | | Redo data.kubernetes.is_deployment
TRAC - deployment.yaml - | | Redo data.kubernetes.is_deployment
TRAC - deployment.yaml - | | | Redo input.kind = "Deployment"
TRAC - deployment.yaml - | Enter data.main.deny
TRAC - deployment.yaml - | | Eval data.kubernetes.is_deployment
TRAC - deployment.yaml - | | Index data.kubernetes.is_deployment (matched 1 rule)
TRAC - deployment.yaml - | | Eval not input.spec.template.spec.securityContext.runAsNonRoot
TRAC - deployment.yaml - | | Enter input.spec.template.spec.securityContext.runAsNonRoot
TRAC - deployment.yaml - | | | Eval input.spec.template.spec.securityContext.runAsNonRoot
TRAC - deployment.yaml - | | | Fail input.spec.template.spec.securityContext.runAsNonRoot
TRAC - deployment.yaml - | | Eval __local7__ = data.main.name
TRAC - deployment.yaml - | | Index __local7__ = data.main.name matched 3 rules)
TRAC - deployment.yaml - | | Eval sprintf("Containers must not run as root in Deployment %s", [__local7__], __local0__)
TRAC - deployment.yaml - | | Eval msg = __local0__
TRAC - deployment.yaml - | | Exit data.main.deny
TRAC - deployment.yaml - | Redo data.main.deny
TRAC - deployment.yaml - | | Redo msg = __local0__
TRAC - deployment.yaml - | | Redo sprintf("Containers must not run as root in Deployment %s", [__local7__], __local0__)
TRAC - deployment.yaml - | | Redo __local7__ = data.main.name
TRAC - deployment.yaml - | | Redo data.kubernetes.is_deployment
TRAC - deployment.yaml - | Enter data.main.deny
TRAC - deployment.yaml - | | Eval data.kubernetes.is_deployment
TRAC - deployment.yaml - | | Index data.kubernetes.is_deployment (matched 1 rule)
TRAC - deployment.yaml - | | Eval not data.main.labels
TRAC - deployment.yaml - | | Enter data.main.labels
TRAC - deployment.yaml - | | | Eval data.main.labels
TRAC - deployment.yaml - | | | Index data.main.labels matched 2 rules)
TRAC - deployment.yaml - | | | Enter data.main.labels
TRAC - deployment.yaml - | | | | Eval input.metadata.labels["app.kubernetes.io/name"]
TRAC - deployment.yaml - | | | | Eval input.metadata.labels["app.kubernetes.io/instance"]
TRAC - deployment.yaml - | | | | Fail input.metadata.labels["app.kubernetes.io/instance"]
TRAC - deployment.yaml - | | | | Redo input.metadata.labels["app.kubernetes.io/name"]
TRAC - deployment.yaml - | | | Enter data.main.labels
TRAC - deployment.yaml - | | | | Eval input.spec.selector.matchLabels.app
TRAC - deployment.yaml - | | | | Eval input.spec.selector.matchLabels.release
TRAC - deployment.yaml - | | | | Fail input.spec.selector.matchLabels.release
TRAC - deployment.yaml - | | | | Redo input.spec.selector.matchLabels.app
TRAC - deployment.yaml - | | | Fail data.main.labels
TRAC - deployment.yaml - | | Eval __local8__ = data.main.name
TRAC - deployment.yaml - | | Index __local8__ = data.main.name matched 3 rules)
TRAC - deployment.yaml - | | Eval sprintf("Deployment %s must provide app/release labels for pod selectors", [__local8__], __local1__)
TRAC - deployment.yaml - | | Eval msg = __local1__
TRAC - deployment.yaml - | | Exit data.main.deny
TRAC - deployment.yaml - | Redo data.main.deny
TRAC - deployment.yaml - | | Redo msg = __local1__
TRAC - deployment.yaml - | | Redo sprintf("Deployment %s must provide app/release labels for pod selectors", [__local8__], __local1__)
TRAC - deployment.yaml - | | Redo __local8__ = data.main.name
TRAC - deployment.yaml - | | Redo data.kubernetes.is_deployment
TRAC - deployment.yaml - | Exit data.main.deny = _
TRAC - deployment.yaml - Redo data.main.deny = _
TRAC - deployment.yaml - | Redo data.main.deny = _
```

</details>


## Plugins

Conftest supports a plugin system to allow for extensions to conftest without editing the codebase. See the [plugin documentation](docs/plugin.md) for more details

## Installation

`conftest` releases are available for Windows, macOS and Linux on the [releases page](https://github.com/instrumenta/conftest/releases).
On Linux and macOS you can download as follows:

```console
$ wget https://github.com/instrumenta/conftest/releases/download/v0.16.0/conftest_0.16.0_Linux_x86_64.tar.gz
$ tar xzf conftest_0.16.0_Linux_x86_64.tar.gz
$ sudo mv conftest /usr/local/bin
```

### Brew

Install with Homebrew on macOS or Linux:

```console
brew tap instrumenta/instrumenta
brew install conftest
```

### Scoop

You can also install using [Scoop](https://scoop.sh/) on Windows:

```console
scoop bucket add instrumenta https://github.com/instrumenta/scoop-instrumenta
scoop install conftest
```

### Docker

`conftest` is also able to be used via Docker. Simply mount your configuration and policy at `/project` and specify the relevant command like so:

```console
$ docker run --rm -v $(pwd):/project instrumenta/conftest test deployment.yaml
FAIL - deployment.yaml - Containers must not run as root in Deployment hello-kubernetes
```

## Inspiration

* [kubetest](https://github.com/garethr/kubetest) was a similar project of mine, using [Skylark](https://docs.bazel.build/versions/master/skylark/language.html)
* [Open Policy Agent](https://www.openpolicyagent.org/) and the Rego query language
* The [helm-opa](https://github.com/eicnix/helm-opa) plugin from [@eicnix](https://github.com/eicnix/) helped with understanding the OPA Go packages
* Tools from the wider infrastructure as code community, in particular rspec-puppet. Lots of my thoughts in [my talk from KubeCon 2017](https://speakerdeck.com/garethr/developer-tooling-for-kubernetes-configurations)

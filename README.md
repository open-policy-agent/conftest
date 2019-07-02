# Conftest

[![CircleCI](https://circleci.com/gh/instrumenta/conftest.svg?style=svg)](https://circleci.com/gh/instrumenta/conftest)

## What

`conftest` is a utility to help you write tests against structured configuration data. For instance you could
write tests for your Kubernetes configurations, or Tekton pipeline definitions, Terraform code, Serverless configs
 or any other structured data.

`conftest` relies on the Rego language from [Open Policy Agent](https://www.openpolicyagent.org/) for writing
the assertions. You can read more about Rego in [How do I write policies](https://www.openpolicyagent.org/docs/how-do-i-write-policies.html)
in the Open Policy Agent documentation.

## Usage

`conftest` allows you to write policies using Open Policy Agent/rego and apply them to one or
more YAML or JSON configuration files. Policies by default should be placed in a directory
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


By default Conftest looks for `deny` and `warn` rules in the `main` namespace. This can be
altered by running `--namespace` or provided on the configuration file.

Assuming you have a Kubernetes deployment in `deployment.yaml` you can run `conftest` like so:

```console
$ conftest test deployment.yaml
deployment.yaml
   Containers must not run as root
   Deployments are not allowed
```

`conftest` can also be used with stdin:

```console
$ cat deployment.yaml | conftest test -
deployment.yaml
   Containers must not run as root
   Deployments are not allowed
```

Note that `conftest` isn't specific to Kubernetes. It will happily let you write tests for any
configuration file using YAML or JSON.

## Examples

You can find examples using various other tools in the `examples ` directory, including:

* [CUE](examples/cue)
* [Kustomize](examples/kustomize)
* [Terraform](examples/terraform)
* [Serverless Framework](examples/serverless)


## Configuration and external policies

Policies are often reusable between different projects, and Conftest supports a mechanism
to specify dependent policies and to download them. The format reuses the [Bundle defined
by Open Policy Agent](https://www.openpolicyagent.org/docs/latest/bundles).

You can download individual policies directly:

```console
conftest pull instrumenta.azurecr.io/test
```

Policies are stored in OCI-compatible registries. You can read more about this idea in
[this post](https://stevelasker.blog/2019/01/25/cloud-native-artifact-stores-evolve-from-container-registries/).

If you have a compatible OCI registry you can also push new policy bundles like so:

```console
conftest push instrumenta.azurecr.io/test
```

Conftest also supports a simple configuration file which can be used to store the
list of dependent bundles and download them in one go. Create a `conftest.toml`
configuration file like the following:

```toml
# You can override the directory in which to store and look for policies
policy = "tests"

# You can overide the namespace which to search for rules
namespace = "conftest"

# An array of individual policies to download. Only the repository
# key is required. If tag is omitted then latest will be used
[[policies]]
repository = "instrumenta.azurecr.io/test"
tag = "latest"
```

With that in place you can use the following command to download all specified policies:

```console
conftest update
```

If you want to download the latest policies and run the tests in one go you can do so with:

```console
conftest test --update <file-to-test>
```


## Debugging queries

When working on more complex queries, or when learning rego, it's useful to see exactly how the policy is
applied. For this purpose you can use the `--trace` flag. This will output a large trace from Open Policy Agent
like the following:


<details>
<summary>Example of trace</summary>

```console
$ conftest test --trace deployment.yaml
deployment.yaml
Enter data.main.deny = _
| Eval data.main.deny = _
| Index data.main.deny = _ (matched 2 rules)
| Enter deny[msg] { data.kubernetes.is_deployment; not input.spec.template.spec.securityContext.runAsNonRoot = true; __local3__ = data.main.name; sprintf("Containers must not run as root in Deployment %s", [__local3__], __local0__); msg = __local0__ }
| | Eval data.kubernetes.is_deployment
| | Index data.kubernetes.is_deployment (matched 1 rule)
| | Enter is_deployment = true { input.kind = "Deployment" }
| | | Eval input.kind = "Deployment"
| | | Exit is_deployment = true { input.kind = "Deployment" }
| | Eval not input.spec.template.spec.securityContext.runAsNonRoot = true
| | | Eval input.spec.template.spec.securityContext.runAsNonRoot = true
| | | Fail input.spec.template.spec.securityContext.runAsNonRoot = true
| | Eval __local3__ = data.main.name
| | Index __local3__ = data.main.name (matched 2 rules)
| | Enter name = __local1__ { true; __local1__ = input.metadata.name }
| | | Eval true
| | | Eval __local1__ = input.metadata.name
| | | Exit name = __local1__ { true; __local1__ = input.metadata.name }
| | Eval sprintf("Containers must not run as root in Deployment %s", [__local3__], __local0__)
| | Eval msg = __local0__
| | Exit deny[msg] { data.kubernetes.is_deployment; not input.spec.template.spec.securityContext.runAsNonRoot = true; __local3__ = data.main.name; sprintf("Containers must not run as root in Deployment %s", [__local3__], __local0__); msg = __local0__ }
| Redo deny[msg] { data.kubernetes.is_deployment; not input.spec.template.spec.securityContext.runAsNonRoot = true; __local3__ = data.main.name; sprintf("Containers must not run as root in Deployment %s", [__local3__], __local0__); msg = __local0__ }
| | Redo msg = __local0__
| | Redo sprintf("Containers must not run as root in Deployment %s", [__local3__], __local0__)
| | Redo __local3__ = data.main.name
| | Redo name = __local1__ { true; __local1__ = input.metadata.name }
| | | Redo __local1__ = input.metadata.name
| | | Redo true
| | Enter name = __local2__ { true; __local2__ = input.metadata.name }
| | | Eval true
| | | Eval __local2__ = input.metadata.name
| | | Exit name = __local2__ { true; __local2__ = input.metadata.name }
| | Redo name = __local2__ { true; __local2__ = input.metadata.name }
| | | Redo __local2__ = input.metadata.name
| | | Redo true
| | Redo data.kubernetes.is_deployment
| | Redo is_deployment = true { input.kind = "Deployment" }
| | | Redo input.kind = "Deployment"
| Enter deny[msg] { data.kubernetes.is_deployment; not data.main.labels; __local4__ = data.main.name; sprintf("Deployment %s must provide app/release labels for pod selectors", [__local4__], __local1__); msg = __local1__ }
| | Eval data.kubernetes.is_deployment
| | Index data.kubernetes.is_deployment (matched 1 rule)
| | Eval not data.main.labels
| | | Eval data.main.labels
| | | Index data.main.labels (matched 1 rule)
| | | Enter labels = true { input.spec.selector.matchLabels.app; input.spec.selector.matchLabels.release }
| | | | Eval input.spec.selector.matchLabels.app
| | | | Eval input.spec.selector.matchLabels.release
| | | | Fail input.spec.selector.matchLabels.release
| | | | Redo input.spec.selector.matchLabels.app
| | | Fail data.main.labels
| | Eval __local4__ = data.main.name
| | Index __local4__ = data.main.name (matched 2 rules)
| | Eval sprintf("Deployment %s must provide app/release labels for pod selectors", [__local4__], __local1__)
| | Eval msg = __local1__
| | Exit deny[msg] { data.kubernetes.is_deployment; not data.main.labels; __local4__ = data.main.name; sprintf("Deployment %s must provide app/release labels for pod selectors", [__local4__], __local1__); msg = __local1__ }
| Redo deny[msg] { data.kubernetes.is_deployment; not data.main.labels; __local4__ = data.main.name; sprintf("Deployment %s must provide app/release labels for pod selectors", [__local4__], __local1__); msg = __local1__ }
| | Redo msg = __local1__
| | Redo sprintf("Deployment %s must provide app/release labels for pod selectors", [__local4__], __local1__)
| | Redo __local4__ = data.main.name
| | Redo data.kubernetes.is_deployment
| Exit data.main.deny = _
Redo data.main.deny = _
| Redo data.main.deny = _
Enter data.main.warn = _
| Eval data.main.warn = _
| Index data.main.warn = _ (matched 1 rule)
| Enter warn[msg] { data.kubernetes.is_service; __local2__ = data.main.name; sprintf("Found service %s but services are not allowed", [__local2__], __local0__); msg = __local0__ }
| | Eval data.kubernetes.is_service
| | Index data.kubernetes.is_service (matched 0 rules)
| | Fail data.kubernetes.is_service
| Exit data.main.warn = _
Redo data.main.warn = _
| Redo data.main.warn = _
   Containers must not run as root in Deployment hello-kubernetes
   Deployment hello-kubernetes must provide app/release labels for pod selectors
```

</details>


## Installation

`conftest` releases are available for Windows, macOS and Linux on the [releases page](https://github.com/instrumenta/conftest/releases).
On Linux and macOS you can probably download as follows:

```console
$ wget https://github.com/instrumenta/conftest/releases/download/v0.7.0/conftest_0.7.0_Linux_x86_64.tar.gz
$ tar xzf conftest_0.7.0_Linux_x86_64.tar.gz
$ sudo mv conftest /usr/local/bin
```

### Brew

If you're on a Mac and using Homebrew you can use:

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

Conftest is also able to be used via Docker. Simply mount your configuration and policy at `/project`
and specify the relevant command like so:

```console
$ docker run --rm -v $(pwd):/project instrumenta/conftest test deployment.yaml
deployment.yaml
   Containers must not run as root in Deployment hello-kubernetes
```

## Inspiration

* [kubetest](https://github.com/garethr/kubetest) was a similar project of mine, using [Skylark](https://docs.bazel.build/versions/master/skylark/language.html)
* [Open Policy Agent](https://www.openpolicyagent.org/) and the Rego query language
* The [helm-opa](https://github.com/eicnix/helm-opa) plugin from [@eicnix](https://github.com/eicnix/) helped with understanding the OPA Go packages
* Tools from the wider instrastructure as code community, in particular rspec-puppet. Lots of my thoughts in [my talk from KubeCon 2017](https://speakerdeck.com/garethr/developer-tooling-for-kubernetes-configurations)
* The code in `pkg/auth` is copied from Oras and will be removed once [this issue](https://github.com/deislabs/oras/issues/98) is resolved


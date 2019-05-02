# Conftest

`conftest` is very much still work-in-progress and could explode at any time.

## What

`conftest` is a utility to help you write tests against structured configuration data. For instance you could
write tests for your Kubernetes configurations, or Tekton pipeline definitions or any other structured data.

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


fail[msg] {
  input.kind = "Deployment"
  not input.spec.template.spec.securityContext.runAsNonRoot = true
  msg = "Containers must not run as root"
}

fail[msg] {
  input.kind = "Deployment"
  not input.spec.selector.matchLabels.app
  msg = "Containers must provide app label for pod selectors"
}
```

Assuming you have a Kubernetes deployment in `deployment.yaml` you can run `conftest` like so:

```console
$ conftest test deployment.yaml
testdata/deployment.yaml
   Containers must not run as root
   Deployments are not allowed
```

`conftest` can also be used with stdin:

```console
$ cat deployment.yaml | conftest test -
testdata/deployment.yaml
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
to specify dependent policies and to download them. Create a `conftest.toml` configuration file like so:

```toml
# You can also override the directory in which to store and look for policies
# policy = "tests"

# An array of individual policies to download. Only the repository
# key is required. If tag is omitted then latest will be used
[[policies]]
repository = "instrumenta.azurecr.io/test"
tag = "latest"
```

This that in place you can use the following command to download all specified policies:

```console
conftest update
```

You can also download individual policies directly, without the need for the configuration
file like so:

```console
conftest pull instrumenta.azurecr.io/test
```

Policies are stored in OCI-compatible registries. You can read more about this idea in
[this post](https://stevelasker.blog/2019/01/25/cloud-native-artifact-stores-evolve-from-container-registries/) from
@SteveLasker. Conftest does not currently provide a mechanism to upload those polices,
for the moment you can use [oras](https://github.com/deislabs/oras/) directly. More
examples coming soon once this code has had a bit more testing.


## Installation

`conftest` releases are available for Windows, macOS and Linux on the [releases page](https://github.com/instrumenta/conftest/releases).
On Linux and macOS you can probably download as follows:

```console
$ wget https://github.com/instrumenta/conftest/releases/download/v0.4.2/conftest_0.4.2_Linux_x86_64.tar.gz
$ tar xzf conftest_0.4.0_Linux_x86_64.tar.gz
$ sudo mv conftest /usr/local/bin
```

More formal packages should be available in the future.


## Inspiration

* [kubtest](https://github.com/garethr/kubetest) was a similar project of mine, using [Starlark](https://docs.bazel.build/versions/master/skylark/language.html)
* [Open Policy Agent](https://www.openpolicyagent.org/) and the Rego query language
* The [helm-opa](https://github.com/eicnix/helm-opa) plugin from [@eicnix](https://github.com/eicnix/) helped with understanding the OPA Go packages
* Tools from the wider instrastructure as code community, in particular rspec-puppet. Lots of my thoughts in [my talk from KubeCon 2017](https://speakerdeck.com/garethr/developer-tooling-for-kubernetes-configurations)
* The code in `pkg/auth` is copied from Oras and will be removed once [this issue](https://github.com/deislabs/oras/issues/98) is resolved


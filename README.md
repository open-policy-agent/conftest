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
$ conftest deployment.yaml
testdata/deployment.yaml
   Containers must not run as root
   Deployments are not allowed
```

`conftest` can also be used with stdin:

```console
$ cat deployment.yaml | conftest -
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

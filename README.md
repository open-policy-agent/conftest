# Conftest

`conftest` is very much still work-in-progress and could explode at any time.

## What

`conftest` is a utility to help you write tests against structured configuration data. For instance you could
write tests for your Kubernetes configurations, or Tekton pipeline definitions or any other structured data.

`conftest` relies on the Rego language from [Open Policy Agent](https://www.openpolicyagent.org/) for writing
the assertions. You can read more about Rego in [How do I write policies](https://www.openpolicyagent.org/docs/how-do-i-write-policies.html)
in the Open Policy Agent documentation.

## Usage

```console
$ conftest --help
Test your configuration files using Open Policy Agent

Usage:
  conftest <file> [file...] [flags]

Flags:
      --fail-on-warn    return a non-zero exit code if only warnings are found
  -h, --help            help for conftest
  -p, --policy string   directory for Rego policy files (default "policy")
      --version         version for conftest
```

```console
$ conftest deployment.yaml
testdata/deployment.yaml
   Containers must not run as root
   Deployments are not allowed
```

## Build

The only way of trying out `conftest` today is to build from source. For that you'll need
a Go toolchain installed. I'll provide binaries at a later date.

```console
$ go build .
```


## Inspiration

* [kubtest](https://github.com/garethr/kubetest) was a similar project of mine, using [Starlark](https://docs.bazel.build/versions/master/skylark/language.html)
* [Open Policy Agent](https://www.openpolicyagent.org/) and the Rego query language
* The [helm-opa](https://github.com/eicnix/helm-opa) plugin from [@eicnix](https://github.com/eicnix/) helped with understanding the OPA Go packages
* Tools from the wider instrastructure as code community, in particular rspec-puppet. Lots of my thoughts in [my talk from KubeCon 2017](https://speakerdeck.com/garethr/developer-tooling-for-kubernetes-configurations)

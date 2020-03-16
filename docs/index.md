# Conftest

Cconftest is a utility to help you write tests against structured configuration data. For instance you could write tests for your Kubernetes configurations, or Tekton pipeline definitions, Terraform code, Serverless configs or any other structured data.

Conftest relies on the Rego language from [Open Policy Agent](https://www.openpolicyagent.org/) for writing the assertions. You can read more about Rego in [How do I write policies](https://www.openpolicyagent.org/docs/how-do-i-write-policies.html) in the Open Policy Agent documentation.

## Usage

Conftest allows you to write policies using Open Policy Agent/rego and apply them to one or
more configuration files. Policies by default should be placed in a directory called `policy` but this can be overridden.

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

2 tests, 0 passed, 0 warnings, 2 failures
```

`conftest` can also be used with stdin:

```console
$ cat deployment.yaml | conftest test -
FAIL - Containers must not run as root
FAIL - Deployments are not allowed

2 tests, 0 passed, 0 warnings, 2 failures
```

Note that Conftest isn't specific to Kubernetes. It will happily let you write tests for any
configuration files. As of today `conftest` supports:

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


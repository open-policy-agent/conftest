# Conftest

Conftest is a utility to help you write tests against structured configuration data. For instance, you could write tests for your Kubernetes configurations, Tekton pipeline definitions, Terraform code, Serverless configs or any other structured data.

Conftest relies on the Rego language from [Open Policy Agent](https://www.openpolicyagent.org/) for writing policies. If you're unsure what exactly a policy is, or unfamiliar with the Rego policy language, the [Policy Language](https://www.openpolicyagent.org/docs/latest/policy-language/) documentation provided by the Open Policy Agent documentation site is a great resource to read.

## Usage

Policies by default should be placed in a directory called `policy`, but this can be overridden with the `--policy` flag.

For instance, save the following as `policy/deployment.rego`:

```rego
package main

deny[msg] {
  input.kind == "Deployment"
  not input.spec.template.spec.securityContext.runAsNonRoot

  msg := "Containers must not run as root"
}

deny[msg] {
  input.kind == "Deployment"
  not input.spec.selector.matchLabels.app

  msg := "Containers must provide app label for pod selectors"
}
```

Conftest looks for `deny`, `violation`, and `warn` rules. Rules can optionally be suffixed with an underscore and an identifier, for example `deny_myrule`.

`violation` rules evaluates the same as `deny` rules, except they support returning structured data errors instead of just strings. See [this issue](https://github.com/open-policy-agent/conftest/pull/243).

By default, Conftest looks for these rules in the `main` namespace, but this can be overriden with the `--namespace` flag or provided in the configuration file. To look in all namespaces, use the `--all-namespaces` flag.

Assuming you have a Kubernetes deployment in `deployment.yaml` you can run Conftest like so:

```console
$ conftest test deployment.yaml
FAIL - deployment.yaml - Containers must not run as root
FAIL - deployment.yaml - Containers must provide app label for pod selectors

2 tests, 0 passed, 0 warnings, 2 failures, 0 exceptions
```

Conftest can also be used with stdin:

```console
$ cat deployment.yaml | conftest test -
FAIL - Containers must not run as root
FAIL - Containers must provide app label for pod selectors

2 tests, 0 passed, 0 warnings, 2 failures, 0 exceptions
```

Note that Conftest isn't specific to Kubernetes. It will happily let you write tests for any configuration files.

As of today Conftest supports:

* CUE
* Dockerfile
* EDN
* HCL and HCL2
* HOCON
* Ignore files (.gitignore, .dockerignore)
* INI
* JSON
* Jsonnet
* TOML
* VCL
* XML
* YAML

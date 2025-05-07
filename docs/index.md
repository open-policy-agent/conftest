# Conftest

Conftest is a utility to help you write tests against structured configuration data. For instance, you could write tests for your Kubernetes configurations, Tekton pipeline definitions, Terraform code, Serverless configs or any other structured data.

Conftest relies on the Rego language from [Open Policy Agent](https://www.openpolicyagent.org/) for writing policies. If you're unsure what exactly a policy is, or unfamiliar with the Rego policy language, the [Policy Language](https://www.openpolicyagent.org/docs/latest/policy-language/) documentation provided by the Open Policy Agent documentation site is a great resource to read.

## Usage

### Evaluating Policies

Policies by default should be placed in a directory called `policy`, but this can be overridden with the `--policy` flag.

For instance, save the following as `policy/deployment.rego`:

```rego
package main

deny contains msg if {
  input.kind == "Deployment"
  not input.spec.template.spec.securityContext.runAsNonRoot

  msg := "Containers must not run as root"
}

deny contains msg if {
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

Conftest supplies a default document with additional contextual information at the `data.conftest` location that can be used in policy evaluation. Currently, the following information is provided:

* `data.conftest.file.name` - The name of the file being evaluated
* `data.conftest.file.dir` - The full directory path of the file being evaluated

Note that Conftest isn't specific to Kubernetes. It will happily let you write tests for any configuration files.

As of today Conftest supports:

* CUE
* CycloneDX
* Dockerfile
* EDN
* Environment files (.env)
* HCL and HCL2
* HOCON
* Ignore files (.gitignore, .dockerignore)
* INI
* JSON
* Jsonnet
* Property files (.properties)
* SPDX
* TextProto (Protocol Buffers)
* TOML
* VCL
* XML
* YAML

### Pre-commit Integration

Conftest can be used as a [pre-commit](https://pre-commit.com/) hook to validate your configuration files before committing them.

To use Conftest with pre-commit, add the following to your `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/open-policy-agent/conftest
    rev: v0.59.0  # Use a specific tag or 'HEAD' for the latest commit
    hooks:
      - id: conftest-test
        args: [--policy, path/to/your/policies]  # Specify your policy directory
      # Optional: Add the verify hook to run policy unit tests
      - id: conftest-verify
        args: [--policy, path/to/your/policies]
```

The `conftest-test` hook validates your configuration files against policies, while the `conftest-verify` hook runs unit tests for your policies.

For more information on pre-commit hooks, refer to the [pre-commit documentation](https://pre-commit.com/).

### Testing/Verifying Policies

When authoring policies, it is helpful to test them. Consult the Rego [testing documentation](https://www.openpolicyagent.org/docs/latest/policy-testing)
for details on testing syntax and approach.

Following the example above, with a policy file in `policy/deployment.rego`, you would create your
tests in `policy/deployment_test.rego` by convention. You can then use `conftest verify` to execute
them and report on the results.

```console
conftest verify --policy ./policy
```

Further documentation can be found using `conftest verify -h`

#### Writing Unit Tests

When writing unit tests, it is common to use the `with` keyword to override the
`input` and `data` documents. For example:

```rego
test_foo if {
  input := {
    "abc": 123,
    "foo": ["bar", "baz"],
  }
  deny with input as input
}
```

However, it can be burdensome to craft the `input` values by hand when the
configurations you are testing are of different formats, especially when they
can be dynamic and their source does not closely align to key-value objects
like Rego requires. A common example is Hashicorp Configuration Language (HCL)
used by Terraform and other products.

To alleviate this issue, conftest provides a builtin function `parse_config`
which takes the parser type and configuration as arguments and parses the
configuration for use in Rego policies. This is the same logic that conftest
uses when testing configurations, only exposed as a Rego function. The example
below shows how to use this to parse an AWS Terraform configuration and use it
in a unit test.

> **TIP:** It is recommended to use the `--show-builtin-errors` flag when
> using the `parse_config`, `parse_config_file`, and `parse_combined_config_files`
> functions. This way errors encountered during parsing will be raised. This
> flag will be enabled by default in a future release.

**deny.rego**

```rego
deny contains msg if {
  proto := input.resource.aws_alb_listener[lb].protocol
  proto == "HTTP"
  msg = sprintf("ALB `%v` is using HTTP rather than HTTPS", [lb])
}
```

**deny_test.rego**

```rego
# "not deny" doesn't work because deny is a set.
# Instead we need to define "no_violations" to be true when `deny` is empty.
empty(value) {
  count(value) == 0
}

no_violations {
  empty(deny)
}

# Now the actual tests start
test_fails_with_http_alb {
  cfg := parse_config("hcl2", `
    resource "aws_alb_listener" "name" {
      protocol = "HTTP"
    }
  `)
  deny["ALB `name` is using HTTP rather than HTTPS"] with input as cfg
}

test_allow_with_alb_https {
  cfg := parse_config("hcl2", `
    resource "aws_alb_listener" "lb_with_https" {
      protocol = "HTTPS"
    }
  `)
  no_violations with input as cfg
}

test_deny_alb_protocol_unspecified {
  cfg := parse_config("hcl2", `
    resource "aws_alb_listener" "lb_with_unspecified_protocol" {
      foo = "bar"
    }
  `)
  no_violations with input as cfg
}
```

For the full list of supported parsers and their names, please refer to the
constants [defined in the parser package](https://pkg.go.dev/github.com/open-policy-agent/conftest/parser#pkg-constants).

If you prefer to have your configuration snippets outside of the Rego unit test
(for syntax highlighting, etc.) you can use the `parse_config_file` builtin. It
accepts the path to the config file as its only parameter and returns the
parsed configuration as a Rego object. The example below shows denying Azure
disks with encryption disabled.

> **NOTE:** The file path argument is relative to the
> location of the Rego unit test file.
>
> **NOTE:** Using this function performs disk I/O which
> can significantly slow down tests.

**deny.rego**

```rego
deny[msg] {
  disk = input.resource.azurerm_managed_disk[name]
  has_field(disk, "encryption_settings")
  disk.encryption_settings.enabled != true
  msg = sprintf("Azure disk `%v` is not encrypted", [name])
}
```

**deny_test.rego**

```rego
test_unencrypted_azure_disk {
  cfg := parse_config_file("unencrypted_azure_disk.tf")
  deny with input as cfg
}
```

**unencrypted_azure_disk.tf**

```hcl
resource "azurerm_managed_disk" "sample" {
  encryption_settings {
    enabled = false
  }
}
```


##### Using `deny_` as a prefix to simplify testing

You may have noticed earlier the weird looking test:

```rego
# Now the actual tests start
test_fails_with_http_alb {
  cfg := parse_config("hcl2", `
    resource "aws_alb_listener" "name" {
      protocol = "HTTP"
    }
  `)
  deny["ALB `name` is using HTTP rather than HTTPS"] with input as cfg
}
```

Specifically, the `deny["ALB ``name`` is using HTTP rather than HTTPS"]` looks a bit strange. The reason we need to do this is we can't just check that any `deny` occurred, we are trying to test that specifically our alb protocol test is working as expected, so we had to match on it's `msg` to make sure we were testing the right rule.

There is an alternative to this, which is to use `deny_` as a prefix, instead of overloading `deny`. For example, we could have instead done:

**deny_v2.rego**

```rego
deny_alb_http[msg] {
  proto := input.resource.aws_alb_listener[lb].protocol
  proto == "HTTP"
  msg = sprintf("ALB `%v` is using HTTP rather than HTTPS", [lb])
}
```

And then we can test specifically that rule with

**deny_v2_test.rego**

```rego
test_fails_with_http_alb {
  cfg := parse_config("hcl2", `
    resource "aws_alb_listener" "name" {
      protocol = "HTTP"
    }
  `)
  deny_alb_http with input as cfg
}
```

This is much more elegant if you have lots of tests and are unit-testing them. Unfortunately you need to do a bit more book-keeping with the `no_violations` rule, but a future feature may make that easier to implement.

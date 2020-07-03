# Exceptions

There might be cases where rules might not apply under certain circumstances. For those occasions, you can use `exceptions`. Exceptions are also written in rego, and allow you to specify policies for when a given `deny` rule does not apply.

The general format of exceptions is as follows:

```rego
exception[rules] {
  # Logic

  rules = ["foo","bar"]
}
```

Inputs matched by the `exception` will be exempted from the rules specified in `rules`, prefixed by `deny_` or `violation_`. The above would provide an exception from for example `deny_foo` or `violation_bar`.

Note that if you specify the empty string, the exception will match *all* rules named just `deny`. It is recommended to use identifiers in your rule names to allow for targeted exceptions.

## Reporting

Exceptions are reported as a separate tally in Conftest's output, so you can detect when exceptions are being made. For example, you might see this summary: `2 tests, 1 passed, 0 warnings, 0 failures, 1 exception`.

## Examples

In the below example, a Kubernetes deployment named `can-run-as-root` will be allowed to run as root, while others will not:

```rego
package main

deny_run_as_root[msg] {
  input.kind = "Deployment"
  not input.spec.template.spec.securityContext.runAsNonRoot = true
  msg = "Containers must not run as root"
}

exception[rules] {
  input.kind = "Deployment"
  input.metadata.name = "can-run-as-root"

  rules = ["run_as_root"]
}
```

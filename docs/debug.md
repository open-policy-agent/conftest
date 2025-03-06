# Debugging policies

When working on more complex queries (or when learning Rego), it's useful to see exactly how the policy is applied. For this purpose you can use the `--trace` flag. This will output a large trace from Open Policy Agent like the following:

```console
$ conftest test --trace deployment.yaml
file: deployment.yaml | query: data.main.deny
TRAC Enter data.main.deny = _
TRAC | Eval data.main.deny = _
TRAC | Index data.main.deny = _ matched 3 rules)
TRAC | Enter data.main.deny
TRAC | | Eval data.kubernetes.is_deployment
TRAC | | Index data.kubernetes.is_deployment (matched 1 rule)
TRAC | | Enter data.kubernetes.is_deployment
TRAC | | | Eval input.kind = "Deployment"
TRAC | | | Exit data.kubernetes.is_deployment
TRAC | | Eval not data.main.required_deployment_labels
TRAC | | Enter data.main.required_deployment_labels
TRAC | | | Eval data.main.required_deployment_labels
TRAC | | | Index data.main.required_deployment_labels (matched 1 rule)
TRAC | | | Enter data.main.required_deployment_labels
TRAC | | | | Eval input.metadata.labels["app.kubernetes.io/name"]
TRAC | | | | Eval input.metadata.labels["app.kubernetes.io/instance"]
TRAC | | | | Fail input.metadata.labels["app.kubernetes.io/instance"]
TRAC | | | | Redo input.metadata.labels["app.kubernetes.io/name"]
TRAC | | | Fail data.main.required_deployment_labels
TRAC | | Eval __local16__ = data.main.name
TRAC | | Index __local16__ = data.main.name matched 4 rules)
TRAC | | Enter data.main.name
TRAC | | | Eval true
TRAC | | | Eval __local9__ = input.metadata.name
TRAC | | | Exit data.main.name
TRAC | | Eval sprintf("%s must include Kubernetes recommended labels: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels", [__local16__], __local5__)
TRAC | | Eval msg = __local5__
TRAC | | Exit data.main.deny
TRAC | Redo data.main.deny
TRAC | | Redo msg = __local5__
TRAC | | Redo sprintf("%s must include Kubernetes recommended labels: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels", [__local16__], __local5__)
TRAC | | Redo __local16__ = data.main.name
TRAC | | Redo data.main.name
TRAC | | | Redo __local9__ = input.metadata.name
TRAC | | | Redo true
TRAC | | Enter data.main.name
TRAC | | | Eval true
TRAC | | | Eval __local10__ = input.metadata.name
TRAC | | | Exit data.main.name
TRAC | | Redo data.main.name
TRAC | | | Redo __local10__ = input.metadata.name
TRAC | | | Redo true
TRAC | | Enter data.main.name
TRAC | | | Eval true
TRAC | | | Eval __local11__ = input.metadata.name
TRAC | | | Exit data.main.name
TRAC | | Redo data.main.name
TRAC | | | Redo __local11__ = input.metadata.name
TRAC | | | Redo true
TRAC | | Enter data.main.name
TRAC | | | Eval true
TRAC | | | Eval __local8__ = input.metadata.name
TRAC | | | Exit data.main.name
TRAC | | Redo data.main.name
TRAC | | | Redo __local8__ = input.metadata.name
TRAC | | | Redo true
TRAC | | Redo data.kubernetes.is_deployment
TRAC | | Redo data.kubernetes.is_deployment
TRAC | | | Redo input.kind = "Deployment"
TRAC | Enter data.main.deny
TRAC | | Eval data.kubernetes.is_deployment
TRAC | | Index data.kubernetes.is_deployment (matched 1 rule)
TRAC | | Eval not input.spec.template.spec.securityContext.runAsNonRoot
TRAC | | Enter input.spec.template.spec.securityContext.runAsNonRoot
TRAC | | | Eval input.spec.template.spec.securityContext.runAsNonRoot
TRAC | | | Fail input.spec.template.spec.securityContext.runAsNonRoot
TRAC | | Eval __local14__ = data.main.name
TRAC | | Index __local14__ = data.main.name matched 4 rules)
TRAC | | Eval sprintf("Containers must not run as root in Deployment %s", [__local14__], __local3__)
TRAC | | Eval msg = __local3__
TRAC | | Exit data.main.deny
TRAC | Redo data.main.deny
TRAC | | Redo msg = __local3__
TRAC | | Redo sprintf("Containers must not run as root in Deployment %s", [__local14__], __local3__)
TRAC | | Redo __local14__ = data.main.name
TRAC | | Redo data.kubernetes.is_deployment
TRAC | Enter data.main.deny
TRAC | | Eval data.kubernetes.is_deployment
TRAC | | Index data.kubernetes.is_deployment (matched 1 rule)
TRAC | | Eval not data.main.required_deployment_selectors
TRAC | | Enter data.main.required_deployment_selectors
TRAC | | | Eval data.main.required_deployment_selectors
TRAC | | | Index data.main.required_deployment_selectors (matched 1 rule)
TRAC | | | Enter data.main.required_deployment_selectors
TRAC | | | | Eval input.spec.selector.matchLabels.app
TRAC | | | | Eval input.spec.selector.matchLabels.release
TRAC | | | | Fail input.spec.selector.matchLabels.release
TRAC | | | | Redo input.spec.selector.matchLabels.app
TRAC | | | Fail data.main.required_deployment_selectors
TRAC | | Eval __local15__ = data.main.name
TRAC | | Index __local15__ = data.main.name matched 4 rules)
TRAC | | Eval sprintf("Deployment %s must provide app/release labels for pod selectors", [__local15__], __local4__)
TRAC | | Eval msg = __local4__
TRAC | | Exit data.main.deny
TRAC | Redo data.main.deny
TRAC | | Redo msg = __local4__
TRAC | | Redo sprintf("Deployment %s must provide app/release labels for pod selectors", [__local15__], __local4__)
TRAC | | Redo __local15__ = data.main.name
TRAC | | Redo data.kubernetes.is_deployment
TRAC | Exit data.main.deny = _
TRAC Redo data.main.deny = _
TRAC | Redo data.main.deny = _
```

## Using trace with other output formats

You can use the `--trace` flag together with any output format. When using `--trace` with formats like `--output=table` or `--output=json`, the trace information will be written to stderr while the formatted output will be written to stdout. This allows you to capture trace information for debugging while still using your preferred output format.

For example:

```console
# Output trace to stderr and table format to stdout
$ conftest test --trace --output=table deployment.yaml

# Capture trace output to a file while viewing table output
$ conftest test --trace --output=table deployment.yaml 2>trace.log
```

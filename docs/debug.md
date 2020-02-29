# Debugging policies

When working on more complex queries (or when learning rego), it's useful to see exactly how the policy is applied. For this purpose you can use the `--trace` flag. This will output a large trace from Open Policy Agent like the following:

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

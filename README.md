# Conftest

[![CircleCI](https://circleci.com/gh/open-policy-agent/conftest.svg?style=svg)](https://circleci.com/gh/open-policy-agent/conftest) [![Netlify](https://api.netlify.com/api/v1/badges/2d928746-3380-4123-b0eb-1fd74ba390db/deploy-status)](https://app.netlify.com/sites/vibrant-villani-65041c/deploys)

Conftest helps you write tests against structured configuration data. Using Conftest you can
write tests for your Kubernetes configuration, Tekton pipeline definitions, Terraform code,
Serverless configs or any other config files.

Conftest uses the Rego language from [Open Policy Agent](https://www.openpolicyagent.org/) for writing
the assertions. You can read more about Rego in [How do I write policies](https://www.openpolicyagent.org/docs/how-do-i-write-policies.html)
in the Open Policy Agent documentation.

Here's a quick example. Save the following as `policy/deployment.rego`:

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

Assuming you have a Kubernetes deployment in `deployment.yaml` you can run Conftest like so:

```console
$ conftest test deployment.yaml
FAIL - deployment.yaml - Containers must not run as root
FAIL - deployment.yaml - Deployments are not allowed

2 tests, 0 passed, 0 warnings, 2 failures
```

Conftest isn't specific to Kubernetes. It will happily let you write tests for any configuration files in a variety of different formats.

See the [documentation](https://www.conftest.dev/) for [installation instructions](https://www.conftest.dev/install/) and
more details about the features. For discussions and questions join us on the [Open Policy Agent Slack](https://slack.openpolicyagent.org/)
in the `#conftest` channel.




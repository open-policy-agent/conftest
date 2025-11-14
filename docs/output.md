# Output options

conftest is designed to be used in CI environments and integrated with other
systems. To support this, the `test` command supports an `--output` flag to
change the format depending on need. You can run `conftest test --help` to see
the available values.

Using the `json` output is available for integration with other tooling that
does not require a specific format. For example:

```json
$ conftest test -p examples/kubernetes/policy/ examples/kubernetes/service.yaml --output json
[
  {
    "filename": "examples/kubernetes/service.yaml",
    "namespace": "main",
    "successes": 4,
    "warnings": [
      {
        "msg": "Found service hello-kubernetes but services are not allowed",
        "metadata": {
          "query": "data.main.warn"
        }
      }
    ]
  }
]
```

For integration with CI systems, additional outputters are available.

- `github`
- `junit`
- `azuredevops`

```
$ conftest test -p examples/kubernetes/policy/ examples/kubernetes/service.yaml --output github
::group::Testing "examples/kubernetes/service.yaml" against 5 policies in namespace "main"
::warning file=examples/kubernetes/service.yaml,line=1::Found service hello-kubernetes but services are not allowed
::notice file=examples/kubernetes/service.yaml,line=1::Number of successful checks: 4
::endgroup::
5 tests, 4 passed, 1 warning, 0 failures, 0 exceptions
```

conftest can also be used to produce SARIF results for integration with systems
that require SBOMs.

```json
$ conftest test -p examples/kubernetes/policy/ examples/kubernetes/service.yaml --output sarif | jq -M .
{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "informationUri": "https://github.com/open-policy-agent/conftest",
          "name": "conftest",
          "rules": [
            {
              "id": "main/warn",
              "shortDescription": {
                "text": "Policy warning"
              },
              "properties": {
                "query": "data.main.warn"
              }
            }
          ]
        }
      },
      "invocations": [
        {
          "executionSuccessful": true,
          "exitCode": 0,
          "exitCodeDescription": "Policy warnings found"
        }
      ],
      "results": [
        {
          "ruleId": "main/warn",
          "ruleIndex": 0,
          "level": "warning",
          "message": {
            "text": "Found service hello-kubernetes but services are not allowed"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "examples/kubernetes/service.yaml"
                }
              }
            }
          ]
        }
      ]
    }
  ]
}
```

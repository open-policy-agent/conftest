# Pre-commit Integration

Conftest can be used as a [pre-commit](https://pre-commit.com/) hook to validate
your configuration files before committing them.

To use Conftest with pre-commit, add the following to your
`.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/open-policy-agent/conftest
    rev: v0.64.0  # Use a specific tag or 'HEAD' for the latest commit
    hooks:
      - id: conftest-test
        args: [--policy, path/to/your/policies]  # Specify your policy directory
      # Optional: Add the verify hook to run policy unit tests
      - id: conftest-verify
        args: [--policy, path/to/your/policies]
```

The `conftest-test` hook validates your configuration files against policies,
while the `conftest-verify` hook runs unit tests for your policies.

Additional hooks are available including `conftest-pull` for downloading
policies and `conftest-fmt` for formatting Rego files. See the
[.pre-commit-hooks.yaml](https://github.com/open-policy-agent/conftest/blob/main/.pre-commit-hooks.yaml)
file for the complete list of available hooks and their configuration options.

For more information on pre-commit hooks, refer to the
[pre-commit documentation](https://pre-commit.com/).

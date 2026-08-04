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

## Documentation Check

The `conftest-doc` hook ensures your policy documentation stays in sync with
your Rego policies. It runs `conftest doc` and if the generated documentation
differs from what's committed, pre-commit will fail. The updated documentation
files are written, so you can simply stage them and re-commit.

By default, the hook documents the `policy` directory. Specifying `args`
replaces the default, so you must include the policy directory path.

```yaml
repos:
  - repo: https://github.com/open-policy-agent/conftest
    rev: v0.64.0
    hooks:
      - id: conftest-doc
        # Uses 'policy' directory by default, or specify your own:
        # args: [path/to/your/policies]
        # To specify an output directory, use -o:
        # args: [-o, docs/, path/to/your/policies]
        # To use a custom template, use --template:
        # args: [--template, path/to/template.md, path/to/your/policies]
        # Combined example:
        # args: [-o, docs/, --template, path/to/template.md, path/to/your/policies]
        # To only run when .rego files in a specific directory change:
        # files: ^path/to/your/policies/.*\.rego$
```

## Additional Hooks

Additional hooks are available including `conftest-pull` for downloading
policies and `conftest-fmt` for formatting Rego files. See the
[.pre-commit-hooks.yaml](https://github.com/open-policy-agent/conftest/blob/main/.pre-commit-hooks.yaml)
file for the complete list of available hooks and their configuration options.

For more information on pre-commit hooks, refer to the
[pre-commit documentation](https://pre-commit.com/).

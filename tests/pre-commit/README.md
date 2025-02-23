# Pre-commit Hook Tests

These tests verify the functionality of Conftest's pre-commit hook integration.

## Test Cases

1. Hook Installation
   - Verifies that the pre-commit hook can be installed successfully

2. Basic Policy Validation
   - Tests single policy validation using the basic example

## Running Tests

The tests are automatically run as part of the project's CI pipeline. To run them locally:

```bash
bats tests/pre-commit/test.bats
```

Note: Requires pre-commit to be installed (`pip install -r requirements-dev.txt`)
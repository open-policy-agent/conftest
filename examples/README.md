# Examples

This folder contains examples of how to use Conftest.

## How to test a manifest against a specific policy in the examples folder

There are various policies with the manifests in the examples folder. They can be tested in a following way:

Run the following command to build the local binary:
```console
make conftest
```

Then, run the following command to test the specific manifest against a specific policy:
```console
./conftest test -p examples/exceptions/policy/ examples/exceptions/deployments.yaml
```

In the above command, we are testing the manifest `examples/exceptions/deployments.yaml` against the policy `examples/exceptions/policy/`.

The `./conftest test` command supports various flags as well for different output formats and configurations. The list of 
supported flags can be displayed with the following command:

```console
./conftest test --help
```

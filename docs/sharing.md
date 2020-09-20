# Sharing policies

Policies are often reusable between different projects, and Conftest supports a mechanism to specify dependent policies as well as download them. The format reuses the [Bundle defined by Open Policy Agent](https://www.openpolicyagent.org/docs/latest/bundles).

You can download individual policies directly:

```console
conftest pull instrumenta.azurecr.io/test
```

The `pull` command also supports other policy locations, such as git or https. Under the hood conftest leverages [go-getter](https://github.com/hashicorp/go-getter) to download policies. For example, to download a policy via https:

```console
conftest pull https://raw.githubusercontent.com/open-policy-agent/conftest/master/examples/compose/policy/deny.rego
```

Policies can be stored in OCI registries that support the Artifact specification. You can read more about this idea in [this post](https://stevelasker.blog/2019/01/25/cloud-native-artifact-stores-evolve-from-container-registries/). Conftest supports storing policies using this mechanism leveraging [ORAS](https://github.com/deislabs/oras).

If you have a compatible OCI registry you can also push new policy bundles like so:

```console
conftest push instrumenta.azurecr.io/test
conftest push 127.0.0.1:5000/test
conftest push <some-other-supported-registry>/test
```

OCI bundles can be pulled as well:

```console
conftest pull instrumenta.azurecr.io/test
conftest pull 127.0.0.1:5000/test
conftest pull oci://<some-other-supported-registry>/test
```

ACR and 127.0.0.1:5000 (The local [Docker Registry](https://github.com/docker/distribution)) are special cases where the URL does not need to be prefixed with the scheme `oci://`, in all other cases the scheme needs to be provided in the URL.

## `--update` flag

If you want to download the latest policies and run the tests in one go, you can do so with the `--update` flag:

```console
conftest test --update <url(s)> <file-to-test>
```

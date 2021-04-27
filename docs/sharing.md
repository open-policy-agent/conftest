# Sharing policies

Policies are often reusable between different projects, and Conftest supports a mechanism to specify dependent policies as well as download them. The format reuses the [Bundle defined by Open Policy Agent](https://www.openpolicyagent.org/docs/latest/bundles).

## Pulling

The `pull` command allows you to download policies using either a URL, a specific protocol (such as `git`), or an [OCI Registry](https://stevelasker.blog/2019/01/25/cloud-native-artifact-stores-evolve-from-container-registries/).

### HTTPS

```console
conftest pull https://raw.githubusercontent.com/open-policy-agent/conftest/master/examples/compose/policy/deny.rego
```

### Git

```console
conftest pull git::https://github.com/<Organization>/<Repository>.git//sub/folder
```

### Git (with access token)

```console
conftest pull git::https://<PersonalAccessToken>@github.com/<Organization>/<Repository>.git//sub/folder
```

### OCI Registry

```console
conftest pull opa.azurecr.io/test
```

See the [go-getter](https://github.com/hashicorp/go-getter) repository for more examples.

## Pushing to an OCI registry

Policies can be stored in OCI registries that support the artifact specification mentioned above. Conftest accomplishes this by leveraging [ORAS](https://github.com/deislabs/oras).

For example, if you have a compatible OCI registry you can push a new policy bundle like so:

```console
conftest push opa.azurecr.io/test
```

## `--update` flag

If you want to download the latest policies and run the tests in one go, you can do so with the `--update` flag:

```console
conftest test --update <url(s)> <file-to-test>
```

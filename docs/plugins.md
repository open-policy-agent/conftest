# Conftest plugins

Conftest provides a plugin feature to allow others to extend the Conftest CLI
without the need to change the Conftest code base. This plugin system was
inspired by the plugin system used in [Helm](https://github.com/helm/helm).

This guide will explain how you can use plugins and how you can create new
plugins.

## Installing a plugin

A plugin can be installed using the `conftest plugin install` command. This
command takes a url and will download the plugin and install it in the plugin
cache.

Conftest adheres to the XDG specification, so the location depends on whether
`XDG_DATA_HOME` or `XDG_DATA_DIRS` is set. Conftest will now search
`XDG_DATA_HOME` or `XDG_DATA_DIRS` for the location of the conftest plugins
cache. The preference order is as follows:

1. XDG_DATA_HOME if set and .conftest/plugins exists within the XDG_DATA_HOME
   dir
1. XDG_DATA_DIRS if set and .conftest/plugins exists within one of the
   XDG_DATA_DIRS
1. ~/.conftest/plugins

Under the hood conftest leverages
[go-getter](https://github.com/hashicorp/go-getter) to download plugins. This
means the following protocols are supported for downloading plugins:

- OCI Registries
- Local Files
- Git
- HTTP/HTTPS
- Mercurial
- Amazon S3
- Google Cloud Storage

For example, to download the Kubernetes Conftest plugin you can execute the
following command:

```console
conftest plugin install github.com/open-policy-agent/conftest//contrib/plugins/kubectl
```

## Using plugins

Once the plugin is installed, Conftest will load all available plugins in the
cache on the start of the next Conftest execution. A plugin will be made in the
Conftest CLI based on the plugin name. For example, to call the kubectl plugin
and audit existing Kubernetes deployments, you can execute the following
command:

```console
conftest kubectl deployment <deployment-id> --policy examples/kubernetes/policy
```

Internally the kubectl plugin calls the kubectl binary to fetch information
about that deployment and passes that information to Conftest. Conftest in turn
executes the Rego policies against the deployment and checks if all policies
pass.

## Developing plugins

A conftest plugin is described by a `plugin.yaml` file. The `plugin.yaml` file
contains metadata about the plugin (e.g. the name of the plugin) and the command
that should be executed when the plugin is triggered.

The `plugin.yaml` field should contain the following information:

- Name: The name of the plugin. This also determines how the plugin will be made
  available in the Conftest CLI. For example, if the plugin is named kubectl,
  you can call the plugin with `conftest kubectl`
- Version: The version of the plugin.
- Usage: A short usage description.
- Description: A long description of the plugin. This is where you could provide
  helpful documentation of your plugin.
- Command: The command that your plugin will execute.

If the plugin contains an executable, that should be stored alongside the
`plugin.yaml`. The relative path to the plugin cache can be given with
`$CONFTEST_PLUGIN_DIR/<my-executable>`.

For example:

```yaml
name: "kubectl"
version: "0.1.0"
usage: conftest kubectl (TYPE[.VERSION][.GROUP] [NAME] | TYPE[.VERSION][.GROUP]/NAME).
description: |-
  A Kubectl plugin for using Conftest to test objects in Kubernetes using Open Policy Agent.
  Usage: conftest kubectl (TYPE[.VERSION][.GROUP] [NAME] | TYPE[.VERSION][.GROUP]/NAME).
command: $CONFTEST_PLUGIN_DIR/kubectl-conftest.sh
```

The plugin is responsible for handling flags and arguments. Any arguments are
passed to the plugin from the conftest command.

Exit codes 1 and 2 are treated as a special exit code in the Conftest CLI. This
indicates a test failure and no error message will be printed. In your plugin
you should return an exit code other than 0, 1, or 2 if your plugin fails for
any reason other than a test failure.

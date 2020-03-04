# Conftest Plugins

Conftest provides a plugin feature to allow others to extend the conftest cli without the need to change the conftest code base. This plugin system was inspired by the plugin system used in [Helm](https://github.com/helm/helm)

This guide will explain how you can use plugins and how you can create new plugins.

## Installing a plugin

A plugin can be installed using the `conftest plugin install` command. This command takes a url and will download the plugin and install it in the plugin cache, located at `~/.conftest/plugins`.

Under the hood conftest leverages [go-getter](https://github.com/hashicorp/go-getter) to download plugins. This means the following protocols are supported for downloading plugins:

- Local Files
- Git
- HTTP/HTTPS
- Mercurial
- Amazon S3
- Google Cloud GCP

For example, to download the example kubernetes conftest plugin you can execute the following command:

```console
conftest plugin install https://github.com/instrumenta/conftest/examples/plugins/kubectl
```

## Using plugins

Once the plugin is installed, `conftest` will load all available plugins in the cache on the start of the next `conftest` execution. A plugin will be made in the conftest cli based on the plugin name. For example, to call the kubectl plugin and audit existing kubernetes deployments, you can execute the following command:

```console
conftest kubectl deployment <deployment-id> --policy examples/kubernetes/policy
```

Internally the kubectl plugin calls the kubectl binary to fetch information about that deployment and passes that information to conftest. Conftest in turn executes the Rego policies against the deployment and checks if all policies pass.

## Developing plugins

A conftest plugin is described by a `plugin.yaml` file. The `plugin.yaml` file contains metadata about the plugin (e.g. the name of the plugin) and the command that should be executed when the plugin is triggered.

The `plugin.yaml` field should contain the following information:

- name: the name of the plugin. This also determines how the plugin will be made available in the conftest cli. E.g. if the plugin is named kubectl, you can call the plugin with `conftest kubectl`
- version: semver version of the plugin.
- usage: a short usage description.
- description: A long description of the plugin. This is where you could provide helpful documentation of your plugin.
- command: The command that your plugin will execute.

If the plugin contains an executable, that should be stored alongside the `plugin.yaml`. The relative path to the plugin cache can be given with `$CONFTEST_PLUGIN_DIR/<my-executable>`.
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

The plugin is responsible for handling flags and arguments. Any arguments are passed to the plugin from the conftest command.

The exit code 1 is treated as a special exit code in the conftest cli. This indicates a test failure and no error message will be printed. In your plugin you should return an exit code other then 0 or 1 if your plugin fails for any reason other then a test failure.

# Installation

Conftest is available for Windows, macOS and Linux on the [releases page](https://github.com/open-policy-agent/conftest/releases). 

On Linux and macOS you can download as follows:

```console
LATEST_VERSION=$(wget -O - "https://api.github.com/repos/open-policy-agent/conftest/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | cut -c 2-)
ARCH=$(arch)
SYSTEM=$(uname)
wget "https://github.com/open-policy-agent/conftest/releases/download/v${LATEST_VERSION}/conftest_${LATEST_VERSION}_${SYSTEM}_${ARCH}.tar.gz"
tar xzf conftest_${LATEST_VERSION}_${SYSTEM}_${ARCH}.tar.gz
sudo mv conftest /usr/local/bin
```

## Brew

Install with Homebrew on macOS or Linux:

```console
brew install conftest
```

## Scoop

You can also install using [Scoop](https://scoop.sh/) on Windows:

```console
scoop install conftest
```

## Docker

Conftest Docker images are also available. Simply mount your configuration and policy at `/project` and specify the relevant command like so:

```console
$ docker run --rm -v $(pwd):/project openpolicyagent/conftest test deployment.yaml
FAIL - deployment.yaml - Containers must not run as root in Deployment hello-kubernetes

1 test, 0 passed, 0 warnings, 1 failure, 0 exceptions
```

> **NOTE:** The instrumenta/conftest image is deprecated and will no longer be updated. Please use the openpolicyagent/conftest image.

## From Source

If you have a working Go environment, you can install conftest from source. It will be installed
to your configured `$GOPATH/bin` folder.

```sh
CGO_ENABLED=0 go install github.com/open-policy-agent/conftest@latest
```

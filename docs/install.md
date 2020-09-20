# Installation

Conftest is available for Windows, macOS and Linux on the [releases page](https://github.com/open-policy-agent/conftest/releases). 

On Linux and macOS you can download as follows:

```console
$ wget https://github.com/open-policy-agent/conftest/releases/download/v0.21.0/conftest_0.21.0_Linux_x86_64.tar.gz
$ tar xzf conftest_0.21.0_Linux_x86_64.tar.gz
$ sudo mv conftest /usr/local/bin
```

## Brew

Install with Homebrew on macOS or Linux:

```console
brew tap instrumenta/instrumenta
brew install conftest
```

## Scoop

You can also install using [Scoop](https://scoop.sh/) on Windows:

```console
scoop bucket add instrumenta https://github.com/instrumenta/scoop-instrumenta
scoop install conftest
```

## Docker

Conftest Docker images are also available. Simply mount your configuration and policy at `/project` and specify the relevant command like so:

```console
$ docker run --rm -v $(pwd):/project instrumenta/conftest test deployment.yaml
FAIL - deployment.yaml - Containers must not run as root in Deployment hello-kubernetes

1 test, 0 passed, 0 warnings, 1 failure, 0 exceptions
```

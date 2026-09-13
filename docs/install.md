# Installation

Conftest is available for Windows, macOS and Linux on the
[releases page](https://github.com/open-policy-agent/conftest/releases).

On Linux and macOS you can download as follows:

```console
LATEST_VERSION=$(wget -O - "https://api.github.com/repos/open-policy-agent/conftest/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | cut -c 2-)
ARCH=$(arch)
SYSTEM=$(uname)
wget "https://github.com/open-policy-agent/conftest/releases/download/v${LATEST_VERSION}/conftest_${LATEST_VERSION}_${SYSTEM}_${ARCH}.tar.gz"
tar xzf conftest_${LATEST_VERSION}_${SYSTEM}_${ARCH}.tar.gz
sudo mv conftest /usr/local/bin
```

## Verifying releases

Every release asset, `checksums.txt` included, and every container image is attested with
[GitHub artifact attestations](https://docs.github.com/en/actions/concepts/security/artifact-attestations).
Each attestation is a SLSA build provenance statement signed keylessly through
Sigstore, so you can prove that an asset was built by this repository's release
workflow from a specific tag, without any keys to distribute.

The commands below pin the signing workflow and the source tag. Both pins
matter: without them a valid attestation produced by any workflow run in the
repository, including one triggered from a pull request, would be accepted.
Make sure the version in the file name or image tag matches the version in
`--source-ref`. The examples below use the Linux x86_64 archive for version
0.70.1; substitute the asset you downloaded, which can be any archive, `.deb`,
or `.rpm` from the release. The commands work the same on Windows with the
`.zip` asset.

Verify a downloaded archive or package with the
[GitHub CLI](https://cli.github.com/) (2.68.0 or newer):

```console
gh attestation verify conftest_0.70.1_Linux_x86_64.tar.gz \
  --repo open-policy-agent/conftest \
  --signer-workflow open-policy-agent/conftest/.github/workflows/release.yaml \
  --source-ref refs/tags/v0.70.1 \
  --deny-self-hosted-runners
```

Verify a container image:

```console
gh attestation verify oci://docker.io/openpolicyagent/conftest:v0.70.1 \
  --repo open-policy-agent/conftest \
  --signer-workflow open-policy-agent/conftest/.github/workflows/release.yaml \
  --source-ref refs/tags/v0.70.1 \
  --deny-self-hosted-runners
```

Image attestations are also pushed to Docker Hub, so
[cosign](https://github.com/sigstore/cosign) (3.0 or newer; older versions
report no signatures) can verify them directly against the registry. Pin the
exact certificate identity and the trigger:

```console
cosign verify docker.io/openpolicyagent/conftest:v0.70.1 \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity https://github.com/open-policy-agent/conftest/.github/workflows/release.yaml@refs/tags/v0.70.1 \
  --certificate-github-workflow-repository open-policy-agent/conftest \
  --certificate-github-workflow-ref refs/tags/v0.70.1 \
  --certificate-github-workflow-trigger push
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

## Mise

You can also install using [Mise](https://github.com/jdx/mise) on Linux/MacOS/Windows:

```console
mise use -g conftest@latest
```


## Docker

Conftest Docker images are also available. Simply mount your configuration and
policy at `/project` and specify the relevant command like so:

```console
$ docker run --rm -v $(pwd):/project openpolicyagent/conftest test deployment.yaml
FAIL - deployment.yaml - Containers must not run as root in Deployment hello-kubernetes

1 test, 0 passed, 0 warnings, 1 failure, 0 exceptions
```

> **NOTE:** The instrumenta/conftest image is deprecated and will no longer be
> updated. Please use the openpolicyagent/conftest image.

## From Source

If you have a working Go environment, you can install conftest from source. It
will be installed to your configured `$GOPATH/bin` folder.

```sh
CGO_ENABLED=0 go install github.com/open-policy-agent/conftest@latest
```

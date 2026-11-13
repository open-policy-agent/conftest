#!/usr/bin/env bash

set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

fake="sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
old=$(grep -oE 'vendorHash = "[^"]*"' flake.nix | cut -d'"' -f2)

sed -i "s|vendorHash = \"${old}\"|vendorHash = \"${fake}\"|" flake.nix
output=$(nix build .#conftest.goModules --no-link 2>&1 || true)
new=$(echo "${output}" | grep -oE 'got: +sha256-[A-Za-z0-9+/=]+' | awk '{print $2}' || true)

if [[ -z "${new}" ]]; then
    sed -i "s|vendorHash = \"${fake}\"|vendorHash = \"${old}\"|" flake.nix
    echo "${output}"
    echo "::error::Unable to determine the vendorHash."
    exit 1
fi

sed -i "s|vendorHash = \"${fake}\"|vendorHash = \"${new}\"|" flake.nix
echo "vendorHash: ${old} -> ${new}"

{
  description = "conftest";

  inputs = {
    nixpkgs = {
      url = "github:nixos/nixpkgs?ref=nixos-unstable";
    };
    flake-utils = {
      url = "github:numtide/flake-utils";
    };
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = nixpkgs.legacyPackages.${system}; in
      {
        devShell = pkgs.mkShell rec {
          packages = with pkgs; [
            bats
            docker
            go
            golangci-lint
            goreleaser
            gnumake
            mdformat
            pipenv
            pre-commit
            ratchet
            regal
          ];
        };
      }
    );

}

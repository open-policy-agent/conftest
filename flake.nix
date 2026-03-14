{
  description = "conftest";

  inputs = {
    nixpkgs = {
      url = "github:nixos/nixpkgs?ref=nixos-unstable";
    };
    flake-utils = {
      url = "github:numtide/flake-utils";
    };
    go-overlay = {
      url = "github:purpleclay/go-overlay";
    };
  };

  outputs = { self, nixpkgs, flake-utils, go-overlay }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs {
        inherit system;
        overlays = [ go-overlay.overlays.default ];
      }; in
      {
        devShell = pkgs.mkShell rec {
          packages = with pkgs; [
            bats
            docker
            (pkgs.go-bin.fromGoMod ./go.mod)
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

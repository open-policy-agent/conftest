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
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ go-overlay.overlays.default ];
        };
        go = pkgs.go-bin.fromGoMod ./go.mod;
        version = self.shortRev or self.dirtyShortRev or "dev";
      in
      {
        packages = rec {
          conftest = (pkgs.buildGoModule.override { inherit go; }) {
            pname = "conftest";
            inherit version;
            src = ./.;
            vendorHash = "sha256-MVOJ8VBw8X3jQ5/nuvf3C89qZE7SUG+Unx4UYimwWcQ=";
            subPackages = [ "." ];
            env.CGO_ENABLED = 0;
            ldflags = [
              "-s"
              "-w"
              "-X github.com/open-policy-agent/conftest/internal/version.Version=${version}"
            ];
            meta = {
              description = "Write tests against structured configuration data using the Open Policy Agent Rego query language";
              homepage = "https://www.conftest.dev";
              license = pkgs.lib.licenses.asl20;
              mainProgram = "conftest";
            };
          };
          default = conftest;
        };

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

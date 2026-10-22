{
  description = "cargo-dep-graph — CLI для визуализации графа зависимостей Cargo";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          name = "cargo-dep-graph-dev";

          packages = with pkgs; [
            go_1_26
            golangci-lint
            gopls
            gotools
            go-tools
            git
          ];

          shellHook = ''
            echo "cargo-dep-graph dev shell"
            echo "Go: $(go version)"
          '';
        };

        packages.default = pkgs.buildGoModule {
          pname = "cargo-dep-graph";
          version = "0.2.0";
          src = ./.;

          subPackages = [ "cmd/cargo-dep-graph" ];

          meta = with pkgs.lib; {
            description = "CLI для визуализации графа зависимостей Cargo";
            license = licenses.mit;
            mainProgram = "cargo-dep-graph";
          };
        };
      }
    );
}

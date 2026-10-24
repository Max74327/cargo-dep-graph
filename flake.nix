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
        go = pkgs.go_1_26;

        # Общие переменные окружения для Go-инструментов в песочнице Nix.
        goEnv = ''
          export GOCACHE=$TMPDIR/go-cache
          export GOPATH=$TMPDIR/go
          export GOLANGCI_LINT_CACHE=$TMPDIR/golangci-lint-cache
          mkdir -p "$GOCACHE" "$GOPATH" "$GOLANGCI_LINT_CACHE"
        '';

        # Копирует исходники в $TMPDIR/src и делает их записываемыми.
        copySrc = ''
          cp -r ${./.} src
          chmod -R +w src
          cd src
        '';

        # Обёртка для "make <target>" в песочнице с нужными инструментами.
        mkCheck =
          target:
          pkgs.runCommand "cargo-dep-graph-${target}"
            {
              nativeBuildInputs = [
                go
                pkgs.golangci-lint
                pkgs.gnumake
              ];
            }
            ''
              ${copySrc}
              ${goEnv}
              make ${target}
              touch $out
            '';
      in
      {
        devShells.default = pkgs.mkShell {
          name = "cargo-dep-graph-dev";
          packages = with pkgs; [
            go
            golangci-lint
            gopls
            gotools
            go-tools
            git
            gnumake
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
          vendorHash = null;
          subPackages = [ "cmd/cargo-dep-graph" ];
          meta = with pkgs.lib; {
            description = "CLI для визуализации графа зависимостей Cargo";
            license = licenses.mit;
            mainProgram = "cargo-dep-graph";
          };
        };

        # Каждая проверка — это make <target> в песочнице.
        checks = {
          fmt = mkCheck "fmt-check";
          vet = mkCheck "vet";
          test = mkCheck "test";
          lint = mkCheck "lint";
          check = mkCheck "check";
        };
      }
    );
}

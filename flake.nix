{
  description = "sqlc-gen-json - sqlc codegen plugin that emits the GenerateRequest as JSON";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = (pkgs.lib.importJSON ./.github/config/release-please-manifest.json).".";

        common = {
          pname = "sqlc-gen-json";
          inherit version;
          src = pkgs.lib.cleanSource ./.;
          subPackages = [ "cmd/sqlc-gen-json" ];
          # proxyVendor=true emits a GOPROXY-style module cache (with all
          # transitive deps populated) instead of a `vendor/` directory.
          # This is what the wasm cross-compile derivation below consumes
          # — go's `go mod vendor` would otherwise prune package paths,
          # because buildGoModule runs it before all imports are visible.
          proxyVendor = true;
          vendorHash = "sha256-NbhVG67LCoSLc0me7OI9WgZb7uICDrf4D5aebW86ggs=";
        };
      in
      {
        packages = {
          default = pkgs.buildGoModule (
            common
            // {
              meta = with pkgs.lib; {
                description = "sqlc plugin that emits the GenerateRequest as JSON";
                license = licenses.mit;
                mainProgram = "sqlc-gen-json";
              };
            }
          );

          # buildGoModule's wrapped Go toolchain overrides GOOS/GOARCH at the
          # toolchain level regardless of env, so cross-compilation to wasip1
          # needs a mkDerivation that calls Go directly. We reuse the module
          # cache that buildGoModule already fetched (common.goModules
          # passthru) and point Go's GOPROXY at it via file:// to avoid
          # network access.
          wasm =
            let
              goModules = (pkgs.buildGoModule common).goModules;
            in
            pkgs.stdenv.mkDerivation {
              pname = "sqlc-gen-json-wasm";
              inherit version;
              src = pkgs.lib.cleanSource ./.;
              nativeBuildInputs = [ pkgs.go ];
              buildPhase = ''
                export HOME=$TMPDIR
                export GOPROXY=file://${goModules}
                export GOFLAGS=-mod=mod
                CGO_ENABLED=0 GOOS=wasip1 GOARCH=wasm \
                  go build -o sqlc-gen-json.wasm \
                  ./cmd/sqlc-gen-json
              '';
              installPhase = ''
                mkdir -p "$out/bin"
                mv sqlc-gen-json.wasm "$out/bin/"
              '';
              doCheck = false;
            };
        };

        devShells.default = pkgs.mkShell {
          name = "sqlc-gen-json";
          packages = [
            pkgs.go
          ];
        };
      }
    );
}

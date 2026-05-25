{
  description = "A basic gomod2nix flake";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.gomod2nix.url = "github:nix-community/gomod2nix";
  inputs.gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
  inputs.gomod2nix.inputs.flake-utils.follows = "flake-utils";

  outputs = { self, nixpkgs, flake-utils, gomod2nix }:
    (flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};

          callPackage = pkgs.callPackage;
        in
        {
          packages.default = callPackage ./. {
            inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
            go = pkgs.go;
          };
          devShells.default = pkgs.mkShell {
            buildInputs = [
              pkgs.go
              pkgs.gcc
              pkgs.libcap
              pkgs.python3
              pkgs.sqlite
              gomod2nix.legacyPackages.${system}.gomod2nix
            ];
            
            FINCH_DB_FILE="/tmp/finch.db";
            FINCH_PORT="9000";
            FINCH_TEMPLATE_DIR="templates";
            FINCH_MEDIA_DIR="media";
            FINCH_SECRET="not-a-real-secret";
            FINCH_ITEMS_PER_PAGE="2";
            FINCH_ALLOW_REGISTRATION="true";
          };
        })
    );
}

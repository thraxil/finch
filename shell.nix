{ pkgs ? (
    let
      inherit (builtins) fetchTree fromJSON readFile;
      inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
    in
    import (fetchTree nixpkgs.locked) {
      overlays = [
        (import "${fetchTree gomod2nix.locked}/overlay.nix")
      ];
    }
  )
, mkGoEnv ? pkgs.mkGoEnv
, gomod2nix ? pkgs.gomod2nix
}:

let
  goEnv = mkGoEnv { pwd = ./.; };
in
pkgs.mkShell {
  packages = [
    goEnv
    gomod2nix
  ];
  FINCH_PORT="7777";
  FINCH_DB_FILE="database.db";
  FINCH_SECRET="secret_for_development";
  FINCH_MEDIA_DIR="media";
  FINCH_ITEMS_PER_PAGE="50";
  FINCH_BASE_URL="http://localhost:7777";
  FINCH_TEMPLATE_DIR="templates";
}

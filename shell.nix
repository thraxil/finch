{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/refs/tags/25.05-pre.tar.gz") {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go
    pkgs.gcc
    pkgs.libcap
    pkgs.python310
    pkgs.sqlite
  ];

  shellHook = ''
  '';

  FINCH_DB_FILE="/tmp/finch.db";
  FINCH_PORT="9000";
  FINCH_TEMPLATE_DIR="templates";
  FINCH_MEDIA_DIR="media";
  FINCH_SECRET="not-a-real-secret";
  FINCH_ITEMS_PER_PAGE="2";
}

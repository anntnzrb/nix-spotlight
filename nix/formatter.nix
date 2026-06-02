{
# Treefmt configuration for Nix and Go formatting.
  pkgs,
}:
pkgs.nixfmt-tree.override {
  runtimeInputs = [
    pkgs.gofumpt
  ];
  settings.formatter.gofumpt = {
    command = "gofumpt";
    options = [ "-l" "-w" ];
    includes = [ "*.go" ];
  };
}

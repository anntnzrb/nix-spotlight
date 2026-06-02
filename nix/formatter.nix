{
  pkgs,
}:
pkgs.nixfmt-tree.override {
  runtimeInputs = [
    pkgs.ruff
    pkgs.gofumpt
  ];
  settings.formatter = {
    ruff-format = {
      command = "ruff";
      options = [ "format" ];
      includes = [ "*.py" ];
    };
    gofumpt = {
      command = "gofumpt";
      options = [ "-l" "-w" ];
      includes = [ "*.go" ];
    };
  };
}

{
  pkgs,
}:
pkgs.nixfmt-tree.override {
  runtimeInputs = [
    pkgs.ruff
  ];
  settings.formatter.ruff-format = {
    command = "ruff";
    options = [ "format" ];
    includes = [ "*.py" ];
  };
}

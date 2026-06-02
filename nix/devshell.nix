{
# Go development shell with linting, formatting, and language-server tools.
  pkgs,
}:
pkgs.mkShell {
  name = "nix-spotlight-dev";

  packages = [
    pkgs.go
    pkgs.golangci-lint
    pkgs.gofumpt
    pkgs.gopls
  ];

  env = {
    GOTOOLCHAIN = "local";
  };

  shellHook = ''
    printf 'go %s | golangci-lint %s\n' \
      "$(go version | awk '{print $3}')" \
      "$(golangci-lint --version | awk '{print $4}')"
  '';
}

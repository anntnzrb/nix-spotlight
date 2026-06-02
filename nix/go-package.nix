{
  pkgs,
  self,
  systems,
}:
pkgs.buildGoModule {
  pname = "nix-spotlight";
  version = "0.2.0-dev";

  src = self;

  vendorHash = null;

  nativeCheckInputs = [ pkgs.golangci-lint ];

  checkPhase = ''
    runHook preCheck
    export GOLANGCI_LINT_CACHE="''${TMPDIR:-/tmp}/golangci-lint-cache"
    mkdir -p "$GOLANGCI_LINT_CACHE"
    golangci-lint run ./...
    go vet ./...
    go test -race -count=1 -shuffle=on ./...
    runHook postCheck
  '';

  meta = {
    description = "macOS Spotlight integration for Nix apps";
    homepage = "https://github.com/anntnzrb/nix-spotlight";
    license = pkgs.lib.licenses.agpl3Only;
    mainProgram = "nix-spotlight";
    platforms = systems;
  };
}

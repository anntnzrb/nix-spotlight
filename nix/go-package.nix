{
# buildGoModule derivation for nix-spotlight.
  pkgs,
  self,
  systems,
}:
let
  version = "0.2.0-dev";
in
pkgs.buildGoModule {
  pname = "nix-spotlight";
  inherit version;

  src = self;

  vendorHash = null;

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
  ];

  nativeCheckInputs = [ pkgs.golangci-lint ];
  nativeBuildInputs = [ pkgs.removeReferencesTo ];
  checkPhase = ''
    runHook preCheck
    export GOLANGCI_LINT_CACHE="''${TMPDIR:-/tmp}/golangci-lint-cache"
    mkdir -p "$GOLANGCI_LINT_CACHE"
    golangci-lint run ./...
    go vet ./...
    go test -race -count=1 -shuffle=on ./...
    runHook postCheck
  '';

  # macOS provides system zoneinfo — the tzdata reference embedded by
  # Go's time package is pure waste in the runtime closure (saves ~2 MiB).
  postFixup = ''
    remove-references-to -t ${pkgs.tzdata} "$out/bin/nix-spotlight"
  '';

  meta = {
    description = "macOS Spotlight integration for Nix apps";
    homepage = "https://github.com/anntnzrb/nix-spotlight";
    license = pkgs.lib.licenses.agpl3Only;
    mainProgram = "nix-spotlight";
    platforms = systems;
  };
}

{
  description = "macOS Spotlight integration for Nix apps";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "aarch64-darwin"
        "x86_64-darwin"
      ];
      forSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      packages = forSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          py = pkgs.python313Packages;
        in
        {
          default = py.buildPythonApplication {
            pname = "nix-spotlight";
            version = "0.1.0";
            pyproject = true;
            src = ./.;

            build-system = [ py.setuptools ];
            nativeCheckInputs = [
              py.mypy
              py.pytest
              py.pytest-cov
              pkgs.ruff
            ];

            checkPhase = ''
              runHook preCheck
              ruff check src/
              mypy src/ --strict
              pytest tests/ -v --cov=nix_spotlight --cov-report=term-missing --cov-fail-under=100
              runHook postCheck
            '';

            meta = {
              description = "macOS Spotlight integration for Nix apps";
              homepage = "https://github.com/anntnzrb/nix-spotlight";
              license = pkgs.lib.licenses.agpl3Only;
              mainProgram = "nix-spotlight";
              platforms = systems;
            };
          };
        }
      );

      homeManagerModules.default =
        {
          pkgs,
          lib,
          config,
          ...
        }:
        let
          cfg = config.programs.nix-spotlight;
        in
        {
          options.programs.nix-spotlight = {
            enable = lib.mkEnableOption "nix-spotlight for Home Manager apps";
            sourceDir = lib.mkOption {
              type = lib.types.str;
              default = "$HOME/Applications/Home Manager Apps";
              description = "Source directory containing .app bundles";
            };
            targetDir = lib.mkOption {
              type = lib.types.str;
              default = "$HOME/Applications/Home Manager Trampolines";
              description = "Target directory for trampolines";
            };
            syncDock = lib.mkOption {
              type = lib.types.bool;
              default = true;
              description = "Whether to sync dock items";
            };
          };

          config = lib.mkIf cfg.enable {
            home.activation.nixSpotlight = lib.hm.dag.entryAfter [ "writeBoundary" ] ''
              ${self.packages.${pkgs.system}.default}/bin/nix-spotlight sync \
                ${lib.optionalString (!cfg.syncDock) "--no-dock"} \
                "${cfg.sourceDir}" \
                "${cfg.targetDir}"
            '';
          };
        };

      darwinModules.default =
        {
          pkgs,
          lib,
          config,
          ...
        }:
        let
          cfg = config.services.nix-spotlight;
        in
        {
          options.services.nix-spotlight = {
            enable = lib.mkEnableOption "nix-spotlight for system apps";
            sourceDir = lib.mkOption {
              type = lib.types.str;
              default = "/Applications/Nix Apps";
              description = "Source directory containing .app bundles";
            };
            targetDir = lib.mkOption {
              type = lib.types.str;
              default = "/Applications/Nix Trampolines";
              description = "Target directory for trampolines";
            };
            syncDock = lib.mkOption {
              type = lib.types.bool;
              default = true;
              description = "Whether to sync dock items";
            };
          };

          config = lib.mkIf cfg.enable {
            system.activationScripts.postActivation.text = ''
              echo "nix-spotlight: syncing trampolines..." >&2
              ${self.packages.${pkgs.system}.default}/bin/nix-spotlight sync \
                ${lib.optionalString (!cfg.syncDock) "--no-dock"} \
                "${cfg.sourceDir}" \
                "${cfg.targetDir}"
            '';
          };
        };
    };
}

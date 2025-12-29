{
  inputs.nixpkgs.url = "https://channels.nixos.org/nixpkgs-unstable/nixexprs.tar.xz";

  outputs =
    { self, nixpkgs }:
    let
      pyproject = builtins.fromTOML (builtins.readFile "${self}/pyproject.toml");
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
            pname = pyproject.project.name;
            version = pyproject.project.version;
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
              description = pyproject.project.description;
              homepage = "https://github.com/anntnzrb/nix-spotlight";
              license = pkgs.lib.licenses.agpl3Only;
              mainProgram = pyproject.project.name;
              platforms = systems;
            };
          };
        }
      );

      homeManagerModules.default =
        { ... }:
        {
          imports = [ "${self}/nix/modules/home-manager.nix" ];
        };

      darwinModules.default =
        { ... }:
        {
          imports = [ "${self}/nix/modules/darwin.nix" ];
        };

      formatter = forSystems (system: nixpkgs.legacyPackages.${system}.nixfmt-tree);
    };
}

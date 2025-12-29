{
  pkgs,
  self,
  systems,
}:
let
  pyproject = builtins.fromTOML (builtins.readFile "${self}/pyproject.toml");
  py = pkgs.python313Packages;
in
py.buildPythonApplication {
  pname = pyproject.project.name;
  version = pyproject.project.version;
  pyproject = true;
  src = self;

  build-system = [ py.setuptools ];
  nativeCheckInputs = [
    pkgs.basedpyright
    py.pytest
    py.pytest-cov
    pkgs.ruff
  ];

  checkPhase = ''
    runHook preCheck
    ruff check src/
    basedpyright src/
    pytest
    runHook postCheck
  '';

  meta = {
    description = pyproject.project.description;
    homepage = "https://github.com/anntnzrb/nix-spotlight";
    license = pkgs.lib.licenses.agpl3Only;
    mainProgram = pyproject.project.name;
    platforms = systems;
  };
}

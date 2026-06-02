{
# nix-spotlight flake for macOS Spotlight trampolines around Nix .app bundles.
  inputs.nixpkgs.url = "https://channels.nixos.org/nixpkgs-unstable/nixexprs.tar.xz";

  outputs =
    { self, nixpkgs }:
    let
      darwinSystems = [
        "aarch64-darwin"
        "x86_64-darwin"
      ];
      allSystems = darwinSystems ++ [ "x86_64-linux" ];
      forDarwin = nixpkgs.lib.genAttrs darwinSystems;
      forAll = nixpkgs.lib.genAttrs allSystems;
    in
    {
      packages = forDarwin (system: let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        go = import "${self}/nix/go-package.nix" { inherit pkgs self; systems = darwinSystems; };
        default = self.packages.${system}.go;
      });

      devShells = forAll (system: {
        default = import "${self}/nix/devshell.nix" {
          pkgs = nixpkgs.legacyPackages.${system};
        };
      });

      homeManagerModules.default = {
        _module.args.self = self;
        imports = [ "${self}/nix/modules/home-manager.nix" ];
      };

      darwinModules.default = {
        _module.args.self = self;
        imports = [ "${self}/nix/modules/darwin.nix" ];
      };

      formatter = forDarwin (
        system: import "${self}/nix/formatter.nix" { pkgs = nixpkgs.legacyPackages.${system}; }
      );
    };
}

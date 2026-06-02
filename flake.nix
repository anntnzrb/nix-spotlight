{
  inputs.nixpkgs.url = "https://channels.nixos.org/nixpkgs-unstable/nixexprs.tar.xz";

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
      packages = forSystems (system: let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        python = import "${self}/nix/package.nix" { inherit pkgs self systems; };
        go = import "${self}/nix/go-package.nix" { inherit pkgs self systems; };
        default = self.packages.${system}.python;
      });

      devShells = forSystems (system: {
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

      formatter = forSystems (
        system: import "${self}/nix/formatter.nix" { pkgs = nixpkgs.legacyPackages.${system}; }
      );
    };
}

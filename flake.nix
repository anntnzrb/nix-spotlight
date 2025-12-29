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
      packages = forSystems (system: {
        default = import "${self}/nix/package.nix" {
          pkgs = nixpkgs.legacyPackages.${system};
          inherit self systems;
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

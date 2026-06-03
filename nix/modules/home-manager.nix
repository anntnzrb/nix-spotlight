{
# Home Manager module for user-scoped nix-spotlight trampoline sync.
  pkgs,
  lib,
  config,
  self,
  ...
}:
let
  shared = import ./shared.nix { inherit lib; };
  cfg = config.programs.nix-spotlight;
in
{
  options.programs.nix-spotlight = (shared.mkOptions {
    defaultSourceDir = "${config.home.homeDirectory}/Applications/Home Manager Apps";
    defaultTargetDir = "${config.home.homeDirectory}/Applications/Home Manager Trampolines";
  }) // {
    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.default;
      description = "The nix-spotlight package to use.";
    };
  };

  config = lib.mkIf cfg.enable {
    home.activation.nixSpotlight = lib.hm.dag.entryAfter [ "writeBoundary" ] ''
      ${shared.mkSyncCommand { inherit cfg; }}
    '';
  };
}

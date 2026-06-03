{
# nix-darwin module for system-scoped nix-spotlight trampoline sync.
  pkgs,
  lib,
  config,
  self,
  ...
}:
let
  shared = import ./shared.nix { inherit lib; };
  cfg = config.services.nix-spotlight;
in
{
  options.services.nix-spotlight = (shared.mkOptions {
    defaultSourceDir = "/Applications/Nix Apps";
    defaultTargetDir = "/Applications/Nix Trampolines";
  }) // {
    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.default;
      description = "The nix-spotlight package to use.";
    };
  };

  config = lib.mkIf cfg.enable {
    system.activationScripts.postActivation.text = ''
      echo "nix-spotlight: syncing trampolines..." >&2
      ${shared.mkSyncCommand { inherit cfg; }}
    '';
  };
}

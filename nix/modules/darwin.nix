{
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
  options.services.nix-spotlight = shared.mkOptions {
    defaultSourceDir = "/Applications/Nix Apps";
    defaultTargetDir = "/Applications/Nix Trampolines";
  };

  config = lib.mkIf cfg.enable {
    system.activationScripts.postActivation.text = ''
      echo "nix-spotlight: syncing trampolines..." >&2
      ${shared.mkSyncCommand { inherit pkgs self cfg; }}
    '';
  };
}

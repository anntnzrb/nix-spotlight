{
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
  options.programs.nix-spotlight = shared.mkOptions {
    defaultSourceDir = "$HOME/Applications/Home Manager Apps";
    defaultTargetDir = "$HOME/Applications/Home Manager Trampolines";
  };

  config = lib.mkIf cfg.enable {
    home.activation.nixSpotlight = lib.hm.dag.entryAfter [ "writeBoundary" ] ''
      ${shared.mkSyncCommand { inherit pkgs self cfg; }}
    '';
  };
}

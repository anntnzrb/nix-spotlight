{
  pkgs,
  lib,
  config,
  self,
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
      ${self.packages.${pkgs.stdenv.hostPlatform.system}.default}/bin/nix-spotlight sync \
        ${lib.optionalString (!cfg.syncDock) "--no-dock"} \
        "${cfg.sourceDir}" \
        "${cfg.targetDir}"
    '';
  };
}
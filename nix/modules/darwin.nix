{
  pkgs,
  lib,
  config,
  self,
  ...
}:
let
  cfg = config.services.nix-spotlight;
in
{
  options.services.nix-spotlight = {
    enable = lib.mkEnableOption "nix-spotlight for system apps";
    sourceDir = lib.mkOption {
      type = lib.types.str;
      default = "/Applications/Nix Apps";
      description = "Source directory containing .app bundles";
    };
    targetDir = lib.mkOption {
      type = lib.types.str;
      default = "/Applications/Nix Trampolines";
      description = "Target directory for trampolines";
    };
    syncDock = lib.mkOption {
      type = lib.types.bool;
      default = true;
      description = "Whether to sync dock items";
    };
  };

  config = lib.mkIf cfg.enable {
    system.activationScripts.postActivation.text = ''
      echo "nix-spotlight: syncing trampolines..." >&2
      ${self.packages.${pkgs.stdenv.hostPlatform.system}.default}/bin/nix-spotlight sync \
        ${lib.optionalString (!cfg.syncDock) "--no-dock"} \
        "${cfg.sourceDir}" \
        "${cfg.targetDir}"
    '';
  };
}
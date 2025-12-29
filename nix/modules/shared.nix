{ lib }:
{
  mkOptions =
    { defaultSourceDir, defaultTargetDir }:
    {
      enable = lib.mkEnableOption "nix-spotlight";
      sourceDir = lib.mkOption {
        type = lib.types.str;
        default = defaultSourceDir;
        description = "Source directory containing .app bundles";
      };
      targetDir = lib.mkOption {
        type = lib.types.str;
        default = defaultTargetDir;
        description = "Target directory for trampolines";
      };
      syncDock = lib.mkOption {
        type = lib.types.bool;
        default = true;
        description = "Whether to sync dock items";
      };
    };

  mkSyncCommand =
    { pkgs, self, cfg }:
    ''
      ${self.packages.${pkgs.stdenv.hostPlatform.system}.default}/bin/nix-spotlight sync \
        ${lib.optionalString (!cfg.syncDock) "--no-dock"} \
        "${cfg.sourceDir}" \
        "${cfg.targetDir}"
    '';
}

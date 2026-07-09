{
  config,
  pkgs,
  lib,
  ...
}:
let
  cfg = config.services.kizuna-ui;

  uiPackage = pkgs.callPackage ../packages/ui.nix { };
in
{

  options.services.kizuna-ui = {
    enable = lib.mkEnableOption "Kizuna UI";

    port = lib.mkOption {
      type = lib.types.port;
      default = 4321;
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.kizuna-ui = {
      description = "Kizuna UI";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      serviceConfig = {
        ExecStart = "${lib.getExe pkgs.static-web-server} --host 0.0.0.0 --port ${toString cfg.port} --root ${uiPackage}/share/kizuna-ui";
        Restart = "always";
        DynamicUser = true;
      };
    };

    networking.firewall.allowedTCPPorts = [
      cfg.port
    ];
  };

}

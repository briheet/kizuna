{
  config,
  pkgs,
  lib,
  ...
}:
let
  cfg = config.services.kizuna-backend;

  backendPackage = pkgs.callPackage ../packages/backend.nix { };
  configFile = pkgs.writeText "kizuna-backend.env" ''
    PORT=${toString cfg.port}
  '';
in
{

  options.services.kizuna-backend = {
    enable = lib.mkEnableOption "Kizuna backend";

    port = lib.mkOption {
      type = lib.types.port;
      default = 8080;
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.kizuna-backend = {
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      environment = {
        PORT = toString cfg.port;
      };

      serviceConfig = {
        ExecStart = "${backendPackage}/bin/backend api --configPath ${configFile}";
        Type = "oneshot";
        RemainAfterExit = true;
        DynamicUser = true;
      };
    };

  };

}

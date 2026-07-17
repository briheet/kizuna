{
  config,
  pkgs,
  lib,
  ...
}:
let
  cfg = config.services.kizuna-workers;

  workersPackage = pkgs.callPackage ../packages/workers.nix { };
  configFile = pkgs.writeText "kizuna-workers.env" ''
    DATABASEURL=${cfg.databaseUrl}
    EMBEDDER_BASE_URL=${cfg.embedderBaseUrl}
    EMBEDDER_MODEL=${cfg.embedderModel}
    CONFLUENCE_HOST=
    CONFLUENCE_MAIL=
    CONFLUENCE_TOKEN=
    DISCORD_TOKEN=
    DISCORD_TOKEN_TYPE=
    GITHUB_TOKEN=
    GITHUB_TOKEN_TYPE=
    SLACK_TOKEN=
    JIRA_HOST=
    JIRA_MAIL=
    JIRA_TOKEN=
  '';
in
{
  options.services.kizuna-workers = {
    enable = lib.mkEnableOption "Kizuna ingestion workers";

    databaseUrl = lib.mkOption {
      type = lib.types.str;
      default = "postgresql://root@127.0.0.1:26257/kizuna?sslmode=disable";
    };

    embedderBaseUrl = lib.mkOption {
      type = lib.types.str;
      default = "http://127.0.0.1:11434";
    };

    embedderModel = lib.mkOption {
      type = lib.types.str;
      default = "nomic-embed-text:v1.5";
    };

  };

  config = lib.mkIf cfg.enable {
    systemd.services.kizuna-workers = {
      description = "Kizuna ingestion workers";
      wantedBy = [ "multi-user.target" ];
      requires = [
        "kizuna-backend-migrate.service"
        "ollama-model-loader.service"
      ];
      after = [
        "kizuna-backend-migrate.service"
        "network.target"
        "ollama-model-loader.service"
      ];

      serviceConfig = {
        ExecStart = "${workersPackage}/bin/workers worker --configPath ${configFile}";
        EnvironmentFile = "%d/workers-secrets.env";
        LoadCredentialEncrypted = [
          "workers-secrets.env:/etc/credstore.encrypted/workers-secrets.env"
        ];
        Type = "simple";
        DynamicUser = true;
        Restart = "always";
        RestartSec = 5;
      };
    };
  };
}

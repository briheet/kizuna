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
    CORS_ALLOWED_ORIGIN=${cfg.corsAllowedOrigin}
    READ_HEADER_TIMEOUT=${toString cfg.readHeaderTimeout}
    READ_TIMEOUT=${toString cfg.readTimeout}
    WRITE_TIMEOUT=${toString cfg.writeTimeout}
    IDLE_TIMEOUT=${toString cfg.idleTimeout}
    DATABASEURL=${cfg.databaseUrl}
    EMBEDDER_BASE_URL=${cfg.embedderBaseUrl}
    EMBEDDER_MODEL=${cfg.embedderModel}
    AI_BASE_URL=${cfg.aiBaseUrl}
    AI_MODEL=${cfg.aiModel}
    AI_MAX_OUTPUT_TOKENS=${toString cfg.aiMaxOutputTokens}
    OPENAI_API_KEY=
  '';

  migrationRunner = pkgs.writeShellApplication {
    name = "kizuna-backend-migrate";
    runtimeInputs = [ config.services.cockroachdb.package ];
    text = ''
      ready=0
      for (( attempt = 1; attempt <= ${toString cfg.migrationMaxAttempts}; attempt++ )); do
        if cockroach sql \
          --url ${lib.escapeShellArg cfg.migrationBootstrapDatabaseUrl} \
          --execute "SELECT 1" >/dev/null 2>&1; then
          ready=1
          break
        fi
        sleep 2
      done

      if [[ "$ready" != "1" ]]; then
        echo "CockroachDB was not ready after ${toString cfg.migrationMaxAttempts} attempts" >&2
        exit 1
      fi

      cockroach sql \
        --url ${lib.escapeShellArg cfg.migrationBootstrapDatabaseUrl} \
        --execute ${lib.escapeShellArg "CREATE DATABASE IF NOT EXISTS ${cfg.migrationDatabaseName};"}

      exec ${backendPackage}/bin/backend migrate \
        --filepath ${lib.escapeShellArg "file://${backendPackage}/share/kizuna-backend/migration"} \
        --dburl ${lib.escapeShellArg cfg.migrationDatabaseUrl}
    '';
  };
in
{

  options.services.kizuna-backend = {
    enable = lib.mkEnableOption "Kizuna backend";

    port = lib.mkOption {
      type = lib.types.port;
      default = 8080;
    };

    openFirewall = lib.mkOption {
      type = lib.types.bool;
      default = false;
      description = "Whether to open the backend API port in the firewall.";
    };

    corsAllowedOrigin = lib.mkOption {
      type = lib.types.str;
      default = "*";
      description = "Value returned in the Access-Control-Allow-Origin header.";
    };

    databaseUrl = lib.mkOption {
      type = lib.types.str;
      default = "postgresql://root@127.0.0.1:26257/kizuna?sslmode=disable";
    };

    migrationDatabaseName = lib.mkOption {
      type = lib.types.strMatching "[A-Za-z_][A-Za-z0-9_]*";
      default = "kizuna";
    };

    migrationBootstrapDatabaseUrl = lib.mkOption {
      type = lib.types.str;
      default = "postgresql://root@127.0.0.1:26257/defaultdb?sslmode=disable";
    };

    migrationDatabaseUrl = lib.mkOption {
      type = lib.types.str;
      default = "cockroachdb://root@127.0.0.1:26257/kizuna?sslmode=disable";
    };

    migrationMaxAttempts = lib.mkOption {
      type = lib.types.ints.positive;
      default = 45;
    };

    embedderBaseUrl = lib.mkOption {
      type = lib.types.str;
      default = "http://127.0.0.1:11434";
    };

    embedderModel = lib.mkOption {
      type = lib.types.str;
      default = "nomic-embed-text:v1.5";
    };

    aiBaseUrl = lib.mkOption {
      type = lib.types.str;
      default = "https://api.openai.com";
    };

    aiModel = lib.mkOption {
      type = lib.types.str;
      default = "gpt-5.4-mini";
    };

    aiMaxOutputTokens = lib.mkOption {
      type = lib.types.ints.between 1 4096;
      default = 700;
    };

    readHeaderTimeout = lib.mkOption {
      type = lib.types.ints.positive;
      default = 5;
    };

    readTimeout = lib.mkOption {
      type = lib.types.ints.positive;
      default = 30;
    };

    writeTimeout = lib.mkOption {
      type = lib.types.ints.positive;
      default = 60;
    };

    idleTimeout = lib.mkOption {
      type = lib.types.ints.positive;
      default = 120;
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.kizuna-backend-migrate = {
      description = "Apply Kizuna CockroachDB migrations";
      wantedBy = [ "multi-user.target" ];
      requires = [ "cockroachdb.service" ];
      after = [
        "cockroachdb.service"
        "network.target"
      ];
      before = [ "kizuna-backend.service" ];

      serviceConfig = {
        ExecStart = lib.getExe migrationRunner;
        Type = "oneshot";
        RemainAfterExit = true;
        DynamicUser = true;
        TimeoutStartSec = 180;
      };
    };

    systemd.services.kizuna-backend = {
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
        ExecStart = "${backendPackage}/bin/backend api --configPath ${configFile}";
        EnvironmentFile = "%d/backend-secrets.env";
        LoadCredentialEncrypted = [
          "backend-secrets.env:/etc/credstore.encrypted/backend-secrets.env"
        ];
        Type = "simple";
        DynamicUser = true;
        Restart = "always";
        RestartSec = 5;
      };
    };

    networking.firewall.allowedTCPPorts = lib.optionals cfg.openFirewall [ cfg.port ];

  };

}

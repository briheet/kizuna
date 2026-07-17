{
  config,
  lib,
  pkgs,
  ...
}:
let
  cfg = config.services.kizuna-nomic;
in
{
  options.services.kizuna-nomic = {
    enable = lib.mkEnableOption "Kizuna Nomic embedding service";

    host = lib.mkOption {
      type = lib.types.str;
      default = "127.0.0.1";
      description = "Address used by the local Ollama embedding API.";
    };

    port = lib.mkOption {
      type = lib.types.port;
      default = 11434;
      description = "Port used by the local Ollama embedding API.";
    };

    model = lib.mkOption {
      type = lib.types.str;
      default = "nomic-embed-text:v1.5";
      description = "Nomic embedding model pulled and served by Ollama.";
    };

    package = lib.mkOption {
      type = lib.types.package;
      default = pkgs.ollama-cpu;
      defaultText = lib.literalExpression "pkgs.ollama-cpu";
      description = "Ollama package used to serve the embedding model.";
    };

    openFirewall = lib.mkOption {
      type = lib.types.bool;
      default = false;
      description = "Whether to expose the embedding API through the host firewall.";
    };
  };

  config = lib.mkIf cfg.enable {
    services.ollama = {
      enable = true;
      inherit (cfg)
        host
        port
        package
        openFirewall
        ;
      loadModels = [ cfg.model ];
    };
  };
}

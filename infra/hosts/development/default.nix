{
  config,
  lib,
  pkgs,
  ...
}:
{

  disabledModules = [ "services/databases/cockroachdb.nix" ];

  imports = [
    ../../modules/backend.nix
    ../../modules/cockroachdb.nix
    ../../modules/nomic.nix
    ../../modules/ui.nix
    ../../modules/workers.nix
  ];

  nixpkgs.config.allowUnfreePredicate =
    pkg:
    builtins.elem (lib.getName pkg) [
      "cockroachdb"
    ];

  # Net
  networking.hostName = "kizuna-dev";

  # ssh
  services.openssh.enable = true;

  users.users.root.openssh.authorizedKeys.keyFiles = [
    ../../secrets/ssh.txt
  ];

  environment.systemPackages = [
    pkgs.curl
    pkgs.gitMinimal
  ];

  services.kizuna-backend = {
    enable = true;
    port = 4000;
    openFirewall = true;
    corsAllowedOrigin = "*";
    embedderBaseUrl = "http://${config.services.kizuna-nomic.host}:${toString config.services.kizuna-nomic.port}";
    embedderModel = config.services.kizuna-nomic.model;
  };

  services.kizuna-nomic.enable = true;

  services.kizuna-workers = {
    enable = true;
    embedderBaseUrl = "http://${config.services.kizuna-nomic.host}:${toString config.services.kizuna-nomic.port}";
    embedderModel = config.services.kizuna-nomic.model;
  };

  services.kizuna-ui = {
    enable = true;
    port = 4321;
  };

  # Services
  services.cockroachdb = {
    enable = true;
    package = pkgs.callPackage ../../packages/cockroachdb.nix { };
    singleNode = true;
    insecure = true;
    stateDirectory = "cockroachdb-v25-4";

    listen = {
      address = "127.0.0.1";
      port = 26257;
    };

    http = {
      address = "127.0.0.1";
      port = 8080;
    };

    openPorts = false;
  };

  system.stateVersion = "26.05";
}

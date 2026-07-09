{
  config,
  lib,
  pkgs,
  ...
}:
let
  secrets = import ../../secrets/development.nix;
in
{

  disabledModules = [ "services/databases/cockroachdb.nix" ];

  imports = [
    ../../modules/backend.nix
    ../../modules/cockroachdb.nix
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
    secrets.publicKeyFile
  ];

  environment.systemPackages = [
    pkgs.curl
    pkgs.gitMinimal
  ];

  services.kizuna-backend = {
    enable = true;
    port = 4000;
  };

  # Services
  services.cockroachdb = {
    enable = true;
    singleNode = true;
    insecure = true;

    listen = {
      address = "0.0.0.0";
      port = 26257;
    };

    http = {
      address = "0.0.0.0";
      port = 8080;
    };

    openPorts = true;

    extraArgs = [
      "--advertise-addr=${secrets.host}:26257"
    ];
  };

  system.stateVersion = "26.05";
}

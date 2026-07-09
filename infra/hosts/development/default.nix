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
    ../../modules/ui.nix
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
  };

  services.kizuna-ui = {
    enable = true;
    port = 4321;
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
      "--advertise-addr=zangetsu:26257"
    ];
  };

  system.stateVersion = "26.05";
}

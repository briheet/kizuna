{
  description = "Kizuna Infrastructure";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    # Disko stuff
    disko.url = "github:nix-community/disko";
    disko.inputs.nixpkgs.follows = "nixpkgs";

    # Deploy-rs stuff
    deploy-rs.url = "github:serokell/deploy-rs";

  };

  outputs =
    {
      self,
      disko,
      deploy-rs,
      nixpkgs,
    }:
    let
      developmentSecrets = import ./secrets/development.nix;
    in
    {

      # Main development machine
      nixosConfigurations.development = nixpkgs.lib.nixosSystem {
        system = "x86_64-linux";
        modules = [
          # dev deployment
          ./hosts/development/default.nix

          # machine
          disko.nixosModules.disko
          ./hardware/configuration.nix
          ./hardware/disk-config.nix
        ];
      };

      # Nixos anywhere setup
      nixosConfigurations.hetzner-cloud = nixpkgs.lib.nixosSystem {
        system = "x86_64-linux";
        modules = [
          disko.nixosModules.disko
          ./hardware/configuration.nix
          ./hardware/disk-config.nix
        ];
      };

      # Deploy-rs node
      deploy.nodes.development = {
        hostname = developmentSecrets.host;
        sshUser = "root";
        remoteBuild = true;

        sshOpts = [
          "-i"
          "~/.ssh/zangetsu"
        ];

        profiles.system = {
          user = "root";
          path = deploy-rs.lib.x86_64-linux.activate.nixos self.nixosConfigurations.development;
        };

        # Cockroach timeout
        confirmTimeout = 300;
        activationTimeout = 600;
      };
      checks = builtins.mapAttrs (system: deployLib: deployLib.deployChecks self.deploy) deploy-rs.lib;

    };

}

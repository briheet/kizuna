# Infra

This is the base note for infra for `Kizuna`
Current infra is based on `nixos-anywhere` and `deploy-rs`.

## VM setup

You need a nixos vm first. `nixos-anywhere` helps with this.
You can start by looking into flake.nix. It currently contains remote VM NixOS setup.

To use this, do this:
1. Copy `infra/secrets/development.example.nix` to `infra/secrets/development.nix`, then set your host IP and public key file there. Keep `development.nix` and the real `.pub` key untracked.
2. Our current cloud config is based on hetzner. You can check it at `flake.nix` at line 27. If using digital ocean or other, please add its config. Take a look here
[nixos anywhere other vm configs](https://github.com/nix-community/nixos-anywhere-examples/blob/main/flake.nix)
3. After adding its config, please check `./hardware/disk-config.nix` to have your particular vm config.
4. Now when done, make sure your system has `nix` installed and `flake` allowed as its experimental.
5. Now hit `nix run github:nix-community/nixos-anywhere -- --flake .#<config_name> --target-host root@<ip>`
For example, `nix run github:nix-community/nixos-anywhere -- --flake .#hetzner-cloud --target-host zangetsu`.
If you have a user setup, please follow this [Target Machine configs](https://nix-community.github.io/nixos-anywhere/quickstart.html#6-connectivity-to-the-target-machine).
6. After connecting to the machine, you would not be able to ssh either ways i.e. `ssh zangetsu` or `ssh root@<ip>`.
Please do `ssh-keygen -R <ssh_setup_name>` or `ssh-keygen -R <ip>`.

You should be able to setup nixos on your vm via this way according to your disk configuration or others. Please edit infra according to your requirements.

For future, when you want to update the machine, hit it with `nixos-rebuild switch --flake .#<config_name> --target-host root@<ip>` or whatever you use.

## Deployment

To deploy your changes to your vm, you can use `deploy-rs`.

Run `nix run github:serokell/deploy-rs -- --skip-checks .#development`
I am using `--skip-checks` here due to being on untrusted.

Also if you read the `flake.nix`, this is an remote build.

## Setup

```nu
~/code/kizuna/infra> ls
╭───┬────────────┬──────┬────────┬────────────────╮
│ # │    name    │ type │  size  │    modified    │
├───┼────────────┼──────┼────────┼────────────────┤
│ 0 │ README.md  │ file │ 2.0 kB │ a minute ago   │
│ 1 │ flake.lock │ file │ 3.4 kB │ 2 hours ago    │
│ 2 │ flake.nix  │ file │ 1.6 kB │ 19 minutes ago │
│ 3 │ hardware   │ dir  │  128 B │ 2 hours ago    │
│ 4 │ hosts      │ dir  │   96 B │ 2 hours ago    │
│ 5 │ modules    │ dir  │  128 B │ 10 minutes ago │
│ 6 │ packages   │ dir  │   96 B │ an hour ago    │
│ 7 │ secrets    │ dir  │   96 B │ 2 hours ago    │
╰───┴────────────┴──────┴────────┴────────────────╯
~/code/kizuna/infra>
```

* hardware: This contains vm's configuration and disk config files.
* hosts: This contains your `hosts/` service setup.
* modules: We make modules of our dependencies and services and then import and use them in our `hosts/`.
* packages: This contains application/package build definitions. These are then consumed by `modules/`.
* secrets: Local machine-specific config lives here. Commit only example files; keep real values such as host IPs and deploy public keys untracked.

# Infra

This is the base note for infra for `Kizuna`
Current infra is based on `nixos-anywhere` and `deploy-rs`.

## VM setup

You need a nixos vm first. `nixos-anywhere` helps with this.
You can start by looking into flake.nix. It currently contains remote VM NixOS setup.

To use this, do this:
1. Put the deploy SSH public key in `infra/secrets/ssh.txt`. Configure the target IP outside the repo with an SSH alias named `zangetsu`.
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
* secrets: Public deploy keys live here. Do not commit private keys, tokens, passwords, or plaintext production secrets.

## Embedding service

The development host runs Ollama on `127.0.0.1:11434` and loads
`nomic-embed-text:v1.5` when the service starts. The backend waits for the model loader
before starting. Check the runtime with:

```sh
systemctl status ollama ollama-model-loader
```

## Database migrations

`kizuna-backend-migrate` waits for CockroachDB, creates the application database
when needed, and applies the SQL migrations packaged with the backend. The
backend starts only after this oneshot completes successfully. Running it again
with `systemctl restart kizuna-backend-migrate` succeeds without changing an
up-to-date database.

```sh
systemctl status cockroachdb kizuna-backend-migrate kizuna-backend kizuna-workers
```

## Required runtime credentials

Create the encrypted dotenv credentials on the VM before the first deployment.

```sh
install -d -m 0700 /etc/credstore.encrypted

read -rsp 'OpenAI API key: ' OPENAI_API_KEY && echo
printf 'OPENAI_API_KEY=%s\n' "$OPENAI_API_KEY" | systemd-creds encrypt --name=backend-secrets.env - /etc/credstore.encrypted/backend-secrets.env
unset OPENAI_API_KEY

read -rsp 'GitHub token: ' GITHUB_TOKEN && echo
printf 'GITHUB_TOKEN=%s\n' "$GITHUB_TOKEN" | systemd-creds encrypt --name=workers-secrets.env - /etc/credstore.encrypted/workers-secrets.env
unset GITHUB_TOKEN

chmod 0600 /etc/credstore.encrypted/backend-secrets.env /etc/credstore.encrypted/workers-secrets.env
```

Run these commands on the target VM so the credentials are bound to that host.
At service start, systemd decrypts each value into its private credential
directory and loads the dotenv file into the service environment. Viper reads
the values through its normal environment override path, and the plaintext
values never enter the repository or Nix store.

For a fine-grained GitHub token, select only the repositories that Kizuna will
index and grant read access to Contents, Issues, and Pull requests. Metadata read
access is included by GitHub.

The backend uses `gpt-5.4-mini` for answer synthesis. GitHub is the only ingestion
provider that needs a credential in the current deployment; the unused Discord,
Slack, Confluence, and Jira settings remain blank. Never add encrypted credential
files or plaintext secret values to the repository.

After deploying, verify the services with:

```sh
systemctl status cockroachdb ollama ollama-model-loader kizuna-backend-migrate kizuna-backend kizuna-workers kizuna-ui
journalctl -u kizuna-backend -u kizuna-workers --since '10 minutes ago'
```

The service starts after the database migration and embedding model are ready.
The development host exposes only the UI on port `4321` and the backend API on
port `4000`; CockroachDB SQL and Admin UI listen on `127.0.0.1` only.

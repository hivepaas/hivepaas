# The Release Installer

HivePaaS has two installers, and neither is for anyone else's server:
`deployment/dev/install.sh` resets the swarm, hardcodes `abc123` for every
secret, deploys the `-dev` images with `*.dev.hivepaas.com` and seeds test data;
`deployment/local/install.sh` runs from a checkout. This design adds the
installer a person runs on their own server:

```bash
curl -fsSL https://raw.githubusercontent.com/hivepaas/hivepaas/main/deployment/release/install.sh | sudo bash
```

It installs the **beta** channel now. Production will be the same script with
the channel set to `stable` once stable images exist; nothing below is specific
to beta except the default.

---

## Decisions

1. **One script, a channel.** `HIVEPAAS_CHANNEL` picks the entry of the release
   info (`beta` by default) and the config file the app is started with
   (`config/config.beta.toml`, later `config.production.toml`).
2. **Images come from the release info**, the same file the app's updater
   reads, so what the installer deploys is what the updater later compares
   against.
3. **Nothing destructive.** The installer never leaves or resets a swarm,
   never removes a service, network or volume, and never regenerates a secret.
   Running it again redeploys the same installation.
4. **Every answer is saved** in `/etc/hivepaas/install.env` (root, `0600`), and a
   second run reads it. The app secret in it is the only key to the encrypted
   data, so the installer never changes it once written.
5. **Interactive by default, silent on request.** Every question has an
   environment variable; `--config FILE` reads them from a file, and `--yes`
   answers yes to installing or upgrading Docker.
6. **Host tuning never fails the install.** Swap and earlyoom are attempted and
   reported; a failure is a warning.
7. **The admin password leaves the configuration** once the app has used it
   (§6).
8. **The dashboard answers on the server's IP too**, until DNS points at it (§5).

## 1. Files

`deployment/release/`:

| file | what |
|---|---|
| `install.sh` | the installer |
| `hivepaas.yaml` | the stack, every host-specific value a `${VAR}` |
| `traefik/dynamic_conf.yml` | the default certificate, as in dev |

The installer downloads the other two from the same ref it came from:
`HIVEPAAS_INSTALL_REF`, default `main`.

## 2. Steps

The script prints the logo, then numbered steps, each `[n/9] What`, in colour
when stdout is a terminal and `NO_COLOR` is unset.

1. **Preflight.**
   - Refuses without root.
   - Detects the distribution from `/etc/os-release`. Supported families:
     Debian/Ubuntu/Raspbian/Mint/Pop (apt), Fedora/RHEL/CentOS/Rocky/Alma/
     Oracle/Amazon Linux (dnf or yum), SLES/openSUSE (zypper), Arch/Manjaro
     (pacman), Alpine (apk). Anything else stops, naming the families.
   - Warns under 2 GB of RAM or 20 GB of free disk; does not stop.
   - Installs what it needs and is missing: `curl`, `openssl`, `jq`,
     `ca-certificates`.
2. **Docker**: at least Engine 29.5 and API 1.54 (§3).
3. **Questions** (§4), skipped for any value `install.env` already holds.
4. **Host memory**: swap and earlyoom (§7).
5. **Swarm.**
   - Not in a swarm: `docker swarm init`. On a host with more than one address
     it passes `--advertise-addr` with the address of the default route, which
     is what `docker swarm init` refuses to guess.
   - In a swarm as a worker: stops - HivePaaS runs on a manager.
   - Labels the node `hivepaas.role=control-plane`.
   - Creates the overlay network `hivepaas_net` if it is missing, with the
     subnet dev uses, `10.11.0.0/16`; when that overlaps a network the host
     has, it lets Docker pick one and says so. Nothing in the app depends on
     the subnet.
6. **Files.**
   - Creates the app data directory, the project data directory, and
     `ssl/certs`, `traefik/etc/dynamic` and `traefik/var/log` under app data.
   - Writes the self-signed certificate the app would otherwise make on first
     boot (`ssl/certs/self-signed.crt|key`): common name the root domain,
     alternative names the root domain, `*.<root domain>` and the app domain.
     The app adopts files it finds there, and Traefik has a certificate from
     its first second rather than its own default one.
   - Downloads the stack and `dynamic_conf.yml`.
7. **Deploy.**
   - Reads `<app data>/system/update/db-volume.env` when an earlier database
     upgrade left one, as `local/install.sh` does.
   - Exports the saved values and runs `docker stack deploy -c hivepaas.yaml
     hivepaas --with-registry-auth`.
   - Sets `--oom-score-adj -500` on the system services, as dev does.
8. **Wait** for the dashboard: polls
   `https://<app domain>/_/ping`, resolved to 127.0.0.1, for up to 5 minutes.
   A timeout is reported with the commands to look at (`docker service ps
   hivepaas_app`, `docker service logs hivepaas_app`), and the script exits 1.
9. **Finish.**
   - Removes the admin password (§6).
   - Prints the logo, the addresses (§5), and the reminders:
     - the certificate is self-signed until one is obtained in the setup, so
       the browser warns and the warning has to be passed once;
     - the server can take 30-60 seconds to answer after a restart;
     - `/etc/hivepaas/install.env` holds the app secret: back it up somewhere
       that is not this server.

## 3. Docker

- **Missing:** asks to install. Yes installs:
  - Debian, Ubuntu, Raspbian, CentOS, Fedora, RHEL: `get.docker.com`;
  - Rocky, Alma, Oracle: Docker's CentOS repository with dnf;
  - Amazon Linux, SLES/openSUSE, Arch, Alpine: the distribution's package.
  Then enables and starts the service.
- **Older than 29.5 or API older than 1.54:** asks to upgrade, the same way.
  No stops: HivePaaS cannot run on it.
- **At least 29.5 but older than the latest release:** asks, defaulting to
  no. The latest version is read from GitHub's releases of `moby/moby`; when
  that fails the question is skipped.
- **After installing or upgrading**, the version is checked again. A
  distribution package still below 29.5 (Amazon Linux, SLES, Alpine can lag)
  stops the install with where to get a newer Docker.
- `--yes` answers yes to the first two and no to the third, unless
  `HIVEPAAS_UPGRADE_DOCKER=true`.

## 4. Questions

Prompts read from `/dev/tty`, not stdin, since `curl | bash` makes stdin the
script. Without a terminal and with a value missing, the installer stops and
lists the variables to set.

| question | variable | default / check |
|---|---|---|
| Admin email | `HIVEPAAS_ADMIN_EMAIL` | an address |
| Admin password | `HIVEPAAS_ADMIN_PASSWORD` | at least 10 characters, typed twice, not echoed |
| App domain | `HIVEPAAS_APP_DOMAIN` | a hostname with at least two labels, e.g. `hivepaas.dev.mydomain.com` |
| Root domain | `HIVEPAAS_ROOT_DOMAIN` | derived and shown for Enter to accept (below) |
| App secret | `HIVEPAAS_APP_SECRET` | Enter generates 64 hex characters; given, at least 32 characters |
| App data directory | `HIVEPAAS_DATA_DIR` | `/var/lib/hivepaas`; an absolute path, not `/` |
| Project data directory | `HIVEPAAS_PROJECT_DATA_DIR` | Enter: `<app data>/project_data`; an absolute path, not `/` |

- **The admin username** is `admin`, not asked.
- **The root domain** is the app domain's last two labels, or last three when
  the last two are a known two-level suffix (`co.uk`, `com.vn`, `com.au`,
  `co.jp`, `com.br`, `co.nz`, `co.za`, `com.sg`, `net.vn`, `org.uk`, and the
  like - a short list in the script). It is shown for the person to accept or
  correct, because no list is complete.
- **Generated, not asked:** `HIVEPAAS_JWT_SECRET` (32 alphanumeric characters),
  the database password, the Redis password and the agent token.
- **Confirmation.** Before anything is changed, the answers are printed - the
  password and the secrets masked - for a last yes.

## 5. The stack

`hivepaas.yaml` is dev's, with these differences:

- **Images** from the release info: `appImage` for app, worker and updater;
  `dbImage`, `redisImage`, `traefikImage`. The agent image is `agentImage` when
  the release info has one, otherwise the app image's repository with `-agent`
  inserted after `hivepaas/hivepaas` and the same tag
  (`hivepaas/hivepaas-dev:0.1.0` gives `hivepaas/hivepaas-agent-dev:0.1.0`), which
  is how the images are built. `HIVEPAAS_AGENT_IMAGE` overrides both.
- **Postgres major** for `PGDATA` is read from the `dbImage` tag.
- **Paths** are the saved directories, never `${PWD}`:
  - app data at `/var/lib/hivepaas` and at `/host<app data>` in app, worker and
    updater, as dev binds `${PWD}/hivepaas`;
  - project data also at `/host<project data>`;
  - Traefik's `etc`, `var/log` and `ssl/certs` under app data.
- **Configuration** by environment: `HP_ENV` and `HP_CONFIG_FILE` from the
  channel; `HP_ROOT_DOMAIN`, `HP_APP_DOMAIN`, `HP_APP_SECRET`,
  `HP_SESSION_JWT_SECRET`, `HP_STORAGE_HOST_DIR`,
  `HP_STORAGE_PROJECT_DATA_HOST_DIR`, `HP_DB_*` with `HP_DB_SSL_MODE=disable`
  (the database is on the stack's own overlay), `HP_CACHE_URL` with the Redis
  password, `HP_AGENT_SECRET_TOKEN`, and the admin's `HP_USER_ADMIN_*`.
- **Redis** is started with `--requirepass`.
- **Traefik**: no `--api.insecure`, no dashboard port, log level `INFO`.
- **No adminer.**
- **The app's routers:**
  - by domain, as dev: `Host(<app domain>)` on `websecure`, and the ACME
    challenge router;
  - by address: `Host(<ip>)` for each IPv4 address the installer found - the
    public one (from `https://ifconfig.io`, skipped on failure) and the host's
    own - on `websecure` with TLS, and on `web` redirecting to HTTPS.
  The address routers and their service are named `x-custom-...`, the marker
  Traefik's label rewrite keeps (`updateSwarmServiceLabels`), so saving the
  app's routing settings later does not remove them. An address that changes
  needs the installer to run again.

## 6. The admin password

The app reads `HP_USER_ADMIN_*` once, on the first boot, when it creates the
admin (`sysInstallationInitData`); afterwards it never looks at them. Kept, the
password would sit in the service definition, readable with `docker service
inspect`, for good.

So once the dashboard answers:
- the password is removed from `install.env`;
- `docker service update --env-rm HP_USER_ADMIN_PASSWORD` runs on app and
  worker, and the installer waits for the dashboard again. That restarts the app
  once, before its address is printed.

The stack writes `HP_USER_ADMIN_PASSWORD: ${HIVEPAAS_ADMIN_PASSWORD:-}`, so a
later run, which no longer has it, deploys it empty.

## 7. Host memory

What `deployment/dev/host-memory.sh` does, folded into the installer, with its
failures made warnings:

- **Swap:** when the host has none and the disk has room, a
  `HIVEPAAS_SWAP_SIZE_MB` file (2048) at `/swapfile`, in `/etc/fstab`;
  `vm.swappiness = 10`. `HIVEPAAS_SWAP=false` skips it.
- **earlyoom:** installed from the distribution where it is packaged - apt,
  dnf (Fedora; RHEL-family only with EPEL, which the installer does not add),
  zypper, pacman, apk - with dev's arguments and avoid list; its config file
  is `/etc/default/earlyoom`, or `/etc/conf.d/earlyoom` on Alpine. Elsewhere it
  is skipped with a note. `HIVEPAAS_EARLYOOM=false` skips it.

earlyoom is Linux-only, as is everything this installer targets. It is worth
installing: without it, a host out of memory stalls for a long time before the
kernel's own OOM killer acts, and every HivePaaS healthcheck fails meanwhile.

## 8. Release info

- Read from `https://raw.githubusercontent.com/hivepaas/hivepaas/<branch>/release.signed.json`,
  `<branch>` being `release`, the branch a non-dev app reads, overridable with
  `HIVEPAAS_RELEASE_BRANCH` (`main` until `release` exists).
- The payload is decoded and its `sha256` checked; a mismatch stops.
- **The signatures are not checked.** The installer, the stack and the release
  info come from the same repository over the same connection; a key embedded
  in the script would be replaced along with anything it protects. The app,
  which keeps its keys in its own binary, checks them as it does now.
- A channel missing from the release info stops the install, naming the
  channels there are.

## 9. Silent install

```bash
sudo HIVEPAAS_ADMIN_EMAIL=... HIVEPAAS_ADMIN_PASSWORD=... HIVEPAAS_APP_DOMAIN=... \
  bash install.sh --yes
sudo bash install.sh --config ./hivepaas.conf --yes
```

A config file is `KEY=VALUE` lines of the variables above, read without being
executed. Command-line environment wins over the file, the file over
`install.env`. `--help` lists every variable.

## 10. Testing

- `shellcheck` on `install.sh`, clean.
- The pure functions - the root domain, the email and password checks, the
  agent image, the Postgres major, reading a config file - sourced by
  `deployment/release/install_test.sh` (`HIVEPAAS_INSTALL_LIB=1` makes the
  script define its functions and return), run by `make test-installer`.
- `docker stack config -c hivepaas.yaml` with a sample `install.env`: the stack
  renders.
- On a fresh Linux server, by the user: an interactive install, a silent one,
  a second run, and the dashboard reached by domain and by IP.

## Not in this design

- **Stable/production images.** Setting the default channel to `stable` is the
  change when they exist.
- **The updater updating the agent.** `sysupdateservice`'s plan has no agent
  step today, so the agent stays on the image the installer deployed.
- **Joining more nodes, uninstalling, IPv6 addresses in the IP routers.**

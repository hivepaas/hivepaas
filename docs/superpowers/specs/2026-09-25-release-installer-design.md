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
   against. A service already running keeps its image: moving images is the
   updater's job, with the migrations a move needs.
3. **Nothing destructive** unless asked. The installer never leaves or resets a
   swarm, never removes a service or network, and never regenerates a secret;
   the one volume it removes is an earlier database the person chose to reset
   (§4).
   Running it again on an installed server does not redeploy (§11).
4. **Settings live where HivePaaS reads them, and nowhere else.** The app
   secret is in `<app data>/hivepaas.toml` (root, `0600`), the app's own
   settings file, where it also keeps a secret it rotates; every other setting
   is the environment of the HivePaaS services. A second run reads them back
   from there. The secret is the only key to the encrypted data, so the
   installer never changes it once written, and `hivepaas.toml` is the one file
   to back up.
5. **Interactive by default, silent on request.** Every question has an
   environment variable; `--config FILE` reads them from a file -
   `deployment/release/install.env` is one to fill in, every setting explained -
   and `--yes` answers yes to installing or upgrading Docker and to the
   confirmations.
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
| `hivepaas.first-boot.yaml` | the app held back (no replicas), deployed with the stack on a first deploy only (§2, step 7) |
| `traefik/dynamic_conf.yml` | the default certificate, as in dev |
| `install.env` | the settings of a silent install, to fill in (§9) |
| `install_test.sh` | the tests of the installer's functions |
| `install_e2e.sh` | a whole install in docker:dind (§10) |

The installer downloads the stack files from the same ref it came from,
`HIVEPAAS_INSTALL_REF` (default `main`), into a temporary directory: it keeps
nothing on the server of its own.

## 2. Steps

The script prints the logo, then numbered steps, each `[n/9] What`, in colour
when stdout is a terminal, `NO_COLOR` is unset and `TERM` is not `dumb`.

1. **Preflight.**
   - Refuses without root.
   - Detects the distribution from `/etc/os-release`, read as data, not
     sourced. Supported families: Debian/Ubuntu/Raspbian and their derivatives
     (apt), Fedora/RHEL/CentOS/Rocky/Alma/Oracle/Amazon Linux (dnf),
     SLES/openSUSE (zypper), Arch/Manjaro (pacman), Alpine (apk). Anything else
     stops, naming the families.
   - Warns when the server is below what is recommended - 4 CPUs, 8 GB of
     RAM, a 40 GB disk under `/var/lib` - with some slack for the sizes a server
     reports; does not stop.
   - Installs what it needs and is missing: `curl`, `openssl`, `jq`,
     `ca-certificates`.
2. **Docker**: at least Engine 29.5 and API 1.54 (§3).
3. **Settings** (§4): read from the environment, `--config` and the HivePaaS
   already on the server; what is still missing is asked. On a first install
   with a terminal and no `--config`, it first says how to install without
   questions (§9). Before anything is changed, the settings are printed - the
   password and the secrets masked - for a last yes.
4. **Host memory**: swap and earlyoom (§7).
5. **Swarm.**
   - Not in a swarm: `docker swarm init --advertise-addr` with the address of
     the default route, which `docker swarm init` refuses to guess on a host
     with more than one.
   - In a swarm as a worker, or a swarm not active (locked, pending): stops -
     HivePaaS runs on a manager.
   - Labels the node `hivepaas.role=control-plane`.
   - Creates the overlay network `hivepaas_net` if it is missing, with the
     subnet dev uses, `10.11.0.0/16`; when a route of the host reaches into
     that subnet, or Docker refuses it, it lets Docker pick one and says so.
     Nothing in the app depends on the subnet.
6. **Files.**
   - Creates the app data directory, the project data directory, and
     `ssl/certs`, `traefik/etc/dynamic` and `traefik/var/log` under app data.
   - Writes the self-signed certificate the app would otherwise make on first
     boot (`ssl/certs/self-signed.crt|key`), anew on every run - one kept from
     an earlier run can have expired, or name another domain: EC P-256, 365
     days, common name the root domain, alternative names the root domain,
     `*.<root domain>` and the app domain. The app adopts files it finds there,
     with their own expiry, unless they are due for renewal; Traefik has a
     certificate from its first second rather than its own default one.
   - Writes `<app data>/hivepaas.toml` when it is not there: the app secret,
     and the security settings the file may also hold, commented out and
     explained. The app rewrites the file, without the comments, when it
     rotates its secret or a security setting is saved in the dashboard.
   - Writes `<app data>/credentials.txt` (`0600`) on every run: the database
     and Redis passwords, the agent token and the JWT secret, each explained,
     as `KEY='value'` lines - readable by a person, and usable as a `--config`
     file. It says not to share it. Never the app secret, nor the admin's
     password.
   - Downloads the stack files, and `dynamic_conf.yml` into Traefik's directory
     only when it is not there: the app writes that directory from its first
     boot on.
7. **Deploy** (a first install; §11 for a second run).
   - Reads `<app data>/system/update/db-volume.env` when an earlier database
     upgrade left one, as `local/install.sh` does, but as data: the two keys it
     knows, checked.
   - Writes `<app data>/first-boot.env` (§6) when the app will make its admin.
   - Exports the settings and runs `docker stack deploy -c hivepaas.yaml -c
     hivepaas.first-boot.yaml hivepaas --with-registry-auth --detach=true`:
     the app deployed with no replicas.
   - One service at a time, `--oom-score-adj -500` on each system service, as
     dev does - `docker stack deploy` drops `oom_score_adj`. Each update
     restarts the service, which is why it happens now, with the app held
     back, and not after its first boot. The database goes first, each waited
     for; an update swarm rolled back stops the install.
   - Migrates the database: the app does not migrate on boot, and the updater
     only on an update. `sql-migrate up` runs in a container of the app image on
     the stack's network, retried until the database takes connections.
   - Starts the app (one replica). Its first boot - the admin, and the request
     for the dashboard's certificate - runs with nothing left to restart it.
8. **Wait** for the dashboard: polls `https://<app domain>/_/ping`, resolved to
   127.0.0.1, for up to 5 minutes. A timeout is reported with the commands to
   look at (`docker service ps hivepaas_app`, `docker service logs
   hivepaas_app`), and the script exits 1. On a run over an installation, whose
   deploy put the OOM priorities back, they are set again here, the app last,
   and the dashboard waited for again.
9. **Finish.**
   - After a first install, waits up to `HIVEPAAS_CERT_WAIT_SECONDS` (20) for
     the dashboard to answer with a certificate the host trusts - `curl`
     without `-k` - which it does once Let's Encrypt has issued the one the
     first boot asked for. A browser that opens the dashboard before then
     keeps its warning until it is restarted; the wait is short because the
     certificate cannot arrive while the domain points elsewhere. Done is
     said either way.
   - Prints the logo, the addresses (§5), and the reminders:
     - whether the dashboard has its certificate; if not, that the browser
       warns until it does, that HivePaaS asks for it once the domain points
       at the server and port 80 is open, the Get started card following it,
       and to quit and reopen the browser once it arrives;
     - the server can take 30-60 seconds to answer after a restart;
     - `<app data>/hivepaas.toml` holds the app secret: back it up somewhere
       that is not this server;
     - the other settings are the services' environment (`docker service
       inspect hivepaas_app`);
     - how to install another server without questions (§9).
   - Every run appends its own lines - steps, results, warnings, errors,
     without colour - under a dated header to `<app data>/install.log`
     (`0600`), once that directory exists. The output of the commands it runs
     is not in it, nor any question, answer or secret: it is the file to share
     when asking for help.

An unexpected failure prints the line it happened on, and that running the
installer again picks up where it stopped.

## 3. Docker

- **Missing:** asks to install. Yes installs:
  - Debian, Ubuntu, Raspbian, CentOS, Fedora, RHEL, Rocky: `get.docker.com`;
  - other derivatives of Ubuntu or Debian (Mint, Pop, Kali...): Docker's apt
    repository, for `UBUNTU_CODENAME`, or `DEBIAN_CODENAME`/`VERSION_CODENAME`;
  - AlmaLinux, Oracle Linux and other RHEL rebuilds: Docker's RHEL repository
    with dnf;
  - Amazon Linux, SLES/openSUSE, Arch, Alpine: the distribution's package.
  Then enables and starts the service (systemd, or OpenRC on Alpine).
- **Older than 29.5 or API older than 1.54:** asks to upgrade, the same way.
  No stops: HivePaaS cannot run on it.
- **At least 29.5 but older than the latest release:** asks, defaulting to
  no. The latest version is the newest in the index of Docker's static builds
  for the architecture (`download.docker.com/linux/static/stable/<arch>/`);
  when that cannot be read the question is skipped.
- **After installing or upgrading**, the version is checked again. A
  distribution package still below 29.5 (Amazon Linux, SLES, Alpine can lag)
  stops the install with where to get a newer Docker.
- `--yes` answers yes to the first two and no to the third, unless
  `HIVEPAAS_UPGRADE_DOCKER=true`. With no terminal and no `--yes`, the first
  two stop the install, and the third is no.

## 4. Settings

Prompts read from `/dev/tty`, not stdin, since `curl | bash` makes stdin the
script. Without a terminal and with a value missing, the installer stops and
lists the variables to set. A value given in the environment or `--config`
that fails its check stops the install, naming the variable and not printing
the value; a typed one that fails is asked again.

| question | variable | default / check |
|---|---|---|
| Admin email | `HIVEPAAS_ADMIN_EMAIL` | an address |
| Admin password | `HIVEPAAS_ADMIN_PASSWORD` | at least 10 characters, typed twice, not echoed |
| App domain | `HIVEPAAS_APP_DOMAIN` | a hostname with at least two labels, e.g. `hivepaas.dev.mydomain.com`; saved lowercased |
| Root domain | `HIVEPAAS_ROOT_DOMAIN` | derived and shown for Enter to accept (below); the app domain or a domain it is under |
| App data directory | `HIVEPAAS_DATA_DIR` | `/var/lib/hivepaas` |
| Project data directory | `HIVEPAAS_PROJECT_DATA_DIR` | Enter: `<app data>/project_data`; not the app data directory or one above it |
| App secret | `HIVEPAAS_APP_SECRET` | not asked when `<app data>/hivepaas.toml` has one; Enter generates 64 hex characters; given, at least 32 characters, no spaces, quotes or backslashes |

- **The admin username** is `admin`, not asked. The admin is not asked for
  once installed.
- **The root domain** is the app domain's last two labels, or last three when
  the last two are a known two-level suffix (`co.uk`, `com.vn`, `com.au`,
  `co.jp`, `com.br`, `co.nz`, `co.za`, `com.sg`, `net.vn`, `org.uk`, and the
  like - a short list in the script). It is shown for the person to accept or
  correct, because no list is complete.
- **A data directory** is an absolute path of letters, digits, `.`, `_` and
  `-`, which Docker can bind as written, and not `/`, a system directory
  (`/etc`, `/usr`, `/proc`, `/run`, `/tmp`, `/var/lib/docker`...) or `/var`,
  `/home`, `/root` themselves. Repeated and trailing slashes are dropped.
- **Generated, not asked:** `HIVEPAAS_JWT_SECRET` (32 alphanumeric characters),
  the database password, the Redis password and the agent token (32 each).
- **Precedence:** the environment, then `--config`, then the HivePaaS already
  on the server: the environment of the `hivepaas_app` service, and the secret
  in `<app data>/hivepaas.toml`. A setting that names where data is or opens it
  - the channel, the domains, the secrets, the directories - cannot change once
  installed: given with another value, it stops the install.
- **What a run cannot guess:**
  - HivePaaS running but `<app data>/hivepaas.toml` missing: the installer stops
    and asks for the file back rather than generate a secret the data was not
    encrypted with.
  - A HivePaaS database volume without HivePaaS services (`docker stack rm`):
    the installer asks at once whether to **keep** or **reset** it - typed out,
    no default, and `--yes` does not answer it; `HIVEPAAS_EXISTING_DB=keep|reset`
    answers it for a silent install, which otherwise stops. Old data can be
    what the person wants back, or what broke the old install.
    - **keep** needs the database's password - `HIVEPAAS_DB_PASSWORD`, or
      `<app data>/credentials.txt` - and `<app data>/hivepaas.toml`. The
      password is tried before anything changes, in a throwaway Postgres of the
      release's image on the volume, with no network. Anything missing, or a
      password that does not open it, stops the install: run it again. The
      admin exists already, so it is not asked for.
    - **reset** deletes the database volumes (`hivepaas_db`, `hivepaas_db_<major>`)
      and `db-volume.env`, and installs afresh, with a new password and a new
      admin.
    - Either waits, first, for the removed stack's containers to let go of the
      volume.

## 5. The stack

`hivepaas.yaml` is dev's, with these differences:

- **Images** from the release info: `appImage` for app, worker and updater;
  `dbImage`, `redisImage`, `traefikImage`. The agent image is `agentImage` when
  the release info has one, otherwise the app image's repository with `-agent`
  inserted after `hivepaas/hivepaas` and the same tag
  (`hivepaas/hivepaas-dev:0.1.0` gives `hivepaas/hivepaas-agent-dev:0.1.0`), which
  is how the images are built. `HIVEPAAS_AGENT_IMAGE` overrides both. A running
  service keeps its image; each service has a variable of its own.
- **Postgres major** for `PGDATA` is read from the `dbImage` tag, unless
  `db-volume.env` names it.
- **Paths** are the saved directories, never `${PWD}`:
  - app data at `/var/lib/hivepaas` and at `/host<app data>` in app, worker and
    updater, as dev binds `${PWD}/hivepaas`;
  - project data also at `/host<project data>` in app and worker;
  - Traefik's `etc`, `var/log` and `ssl/certs` under app data.
- **Configuration** by environment, the same for app, worker, updater and
  agent - each loads the configuration and reaches the database and the
  cache - except the app secret, which no service is given: app, worker and
  updater read it from `hivepaas.toml` at `/var/lib/hivepaas`, and the agent,
  which runs on every node and decrypts nothing, needs none: `HP_ENV` and `HP_CONFIG_FILE` from the channel; `HP_ROOT_DOMAIN`,
  `HP_APP_DOMAIN`, `HP_SESSION_JWT_SECRET` (not for the agent),
  `HP_STORAGE_HOST_DIR`, `HP_STORAGE_PROJECT_DATA_HOST_DIR`, `HP_DB_*` with
  `HP_DB_SSL_MODE=disable` (the database is on the stack's own overlay),
  `HP_CACHE_URL` with the Redis password, `HP_AGENT_SECRET_TOKEN`, and
  `HP_HTTP_SERVER_CORS_ALLOW_ORIGINS=["*"]`, as dev: the app does not start
  without an origin list, the dashboard is served by the app itself at the
  domain and at the addresses, and its session cookies are `SameSite=Lax`.
- **Redis** is started with `--requirepass`.
- **Traefik**: no `--api.insecure`, no dashboard port, log level `INFO`.
- **No adminer.**
- **The app's routers:**
  - by domain, as dev: `Host(<app domain>)` on `websecure`, and the ACME
    challenge router;
  - by address: `Host(<ip>)` for each IPv4 address the installer found - the
    public one (from `https://ifconfig.io`, skipped on failure) and the host's
    own on the default route - on `websecure` with TLS, and on `web`
    redirecting to HTTPS. With no address found, the rule is
    ``Host(`127.0.0.1`)``, so the router still has one.
  The address routers, the ACME router and the service they share are named
  `x-custom-...`, the marker Traefik's label rewrite keeps
  (`updateSwarmServiceLabels`), so saving the app's routing settings later
  does not remove them. Each names its service: with two services on the app,
  Traefik would not pick one for a router that leaves it out.

## 6. The admin password: `first-boot.env`

The app reads `HP_USER_ADMIN_*` once, on the first boot, when it creates the
admin (`sysInstallationInitData`, in the app or the worker, whichever starts
first); afterwards it never looks at them. In a service's environment the
password would be readable with `docker service inspect`, and taking it out
means a service update, which restarts the service - right when its first boot
asks for the dashboard's certificate.

So the installer writes the account to `<app data>/first-boot.env` (`0600`),
never to a service: `HP_*` variables, one `KEY=VALUE` a line, the value as it
stands after the first `=` - no quotes, no escapes. The app reads the file with
its configuration, under the service's own environment, and deletes it once
the installation's data exists; the installer deletes one the app could not.
The file is there only until the first boot has run, which is how a later run
tells an install that has not finished, and where it reads the account back.

The file is for values the first boot uses once. What every boot needs belongs
in `hivepaas.toml` (the managed settings): `first-boot.env` does not survive
the first boot.

## 7. Host memory

What `deployment/dev/host-memory.sh` does, folded into the installer, with its
failures made warnings:

- **Swap:** when the host has none, `/swapfile` does not exist and the disk
  has room, a `HIVEPAAS_SWAP_SIZE_MB` file (2048) at `/swapfile`, in
  `/etc/fstab`; `vm.swappiness = 10`. `HIVEPAAS_SWAP=false` skips both.
- **earlyoom:** installed from the distribution where it is packaged - apt,
  dnf (Fedora; RHEL-family only with EPEL, which the installer does not add),
  zypper, pacman, apk (with `earlyoom-openrc`) - with dev's arguments and avoid
  list. Its configuration is `EARLYOOM_ARGS` in `/etc/default/earlyoom` (apt,
  dnf, pacman) or `/etc/sysconfig/earlyoom` (zypper), and OpenRC's variables in
  `/etc/conf.d/earlyoom` on Alpine. Where it is not packaged it is skipped with
  a note. `HIVEPAAS_EARLYOOM=false` skips it.

earlyoom is Linux-only, as is everything this installer targets. It is worth
installing: without it, a host out of memory stalls for a long time before the
kernel's own OOM killer acts, and every HivePaaS healthcheck fails meanwhile.

## 8. Release info

- Read from `https://raw.githubusercontent.com/hivepaas/hivepaas/<branch>/release.signed.json`,
  `<branch>` being `release`, the branch a non-dev app reads, overridable with
  `HIVEPAAS_RELEASE_BRANCH` (`main` until `release` exists). The error when it
  cannot be read names the variable.
- The payload is decoded and its `sha256` checked; a mismatch stops.
- **The signatures are not checked.** The installer, the stack and the release
  info come from the same repository over the same connection; a key embedded
  in the script would be replaced along with anything it protects. The app,
  which keeps its keys in its own binary, checks them as it does now.
- A channel missing from the release info stops the install, naming the
  channels there are.
- It is read for a first install and for a redeploy, before the settings are
  confirmed, so that a wrong branch stops the install before it changes
  anything.

## 9. Silent install

```bash
curl -fsSLO https://raw.githubusercontent.com/hivepaas/hivepaas/main/deployment/release/install.env
# fill it in
sudo bash install.sh --config install.env --yes
# or
sudo HIVEPAAS_ADMIN_EMAIL=... HIVEPAAS_ADMIN_PASSWORD=... HIVEPAAS_APP_DOMAIN=... \
  bash install.sh --yes
```

`deployment/release/install.env` has the three required settings, empty, and
every other one commented out with what it does and its default. `--help`, the
start of an interactive first install and the end of every install point to
it.

A config file is `KEY=VALUE` lines of the variables above, read without being
executed: `#` comments, blank lines, `export ` and single or double quotes are
allowed, CRLF line ends are taken, and a name that is not `HIVEPAAS_...` is
ignored with a warning. Command-line environment wins over the file, the file
over what the installed services say. `--help` lists every variable.

For tests only: `HIVEPAAS_INSTALL_LIB=1` defines the functions and stops;
`HIVEPAAS_INSTALL_FILES_DIR` copies the stack files from a directory;
`HIVEPAAS_RELEASE_URL` reads the release info from a URL (`file://` works);
`HIVEPAAS_TTY` reads answers from a file; `HIVEPAAS_WAIT_SECONDS` changes the
300-second waits.

## 10. Testing

- `shellcheck` on `install.sh` and the test scripts, clean.
- The pure functions - the checks, the root domain, the settings files, the
  release info, the images, the addresses, the distribution helpers - and the
  questions (answered through `HIVEPAAS_TTY`) in
  `deployment/release/install_test.sh`, under macOS's bash 3.2 as under bash
  5, run by `make test-installer`.
- `docker stack config -c hivepaas.yaml [-c hivepaas.first-boot.yaml]` with
  sample values: the stack renders, the app held back by the first-boot file
  and no service given the admin's account; every `${VAR}` it uses is one the
  installer exports.
- `make test-installer-e2e` (`install_e2e.sh`): in a docker:dind container, a
  silent install from the `install.env` template with this checkout's release
  info; the dashboard by domain and by address, HTTP redirected to HTTPS, the
  admin signing in, the admin's account in no service and `first-boot.env`
  deleted, the app started once and no service restarted after its first
  boot, the certificate's state told after its wait, the app secret
  in `hivepaas.toml` and in no service, the agent running without it, OOM
  priorities, `credentials.txt` and `install.log`; a second run that changes
  no service and makes the self-signed certificate anew; another app secret stopping; a lost `hivepaas.toml` stopping;
  the address route following a change; `--redeploy`; and after
  `docker stack rm`, a silent run stopping until told to keep or reset the
  database, a wrong password stopping `keep`, `keep` with `credentials.txt`
  bringing the old admin back, and `reset` starting over with a new one. `HIVEPAAS_E2E_IMAGES` loads images built from the
  checkout, for an app change the release's images do not have yet. On a machine that is not amd64 the HivePaaS images run
  emulated and the stack is deployed without resolving images.
- On a fresh Linux server, by the user: an interactive install, a silent one,
  a second run, and the dashboard reached by domain and by IP.

## 11. Running it again

- **Installed** (the `hivepaas_app` service there, and no
  `<app data>/first-boot.env`): nothing is asked. The host is
  checked again (Docker, memory, swarm, files), the routes by address follow
  an address that changed - a label update, which restarts nothing, and left
  alone when the public address cannot be looked up - and the system services
  get their OOM priority back if something reset it. The stack is not
  redeployed: HivePaaS changes its own services at runtime (Traefik's
  arguments, the worker and updater replicas, routing labels), and a deploy
  would put them back as the stack file has them.
- **`--redeploy`** deploys the stack again, after a warning saying so; the
  running images stay. The deploy restarts the database with the app, and an
  app update swarm rolls back meanwhile is deployed once more.
- **Unfinished** (the stack deployed, `first-boot.env` still there): the OOM
  priorities, the migrations and the app's start, then the wait and the
  finish, without a new deploy.

## Not in this design

- **Stable/production images.** Setting the default channel to `stable` is the
  change when they exist.
- **The updater updating the agent.** `sysupdateservice`'s plan has no agent
  step today, so the agent stays on the image the installer deployed.
- **Joining more nodes, uninstalling, IPv6 addresses in the IP routers.**
- **The database in the app data directory.** Postgres is on the named volume
  `hivepaas_db`, under Docker's own directory. Moving it to
  `<app data>/db/<major>` needs the updater's major upgrade to follow a bind
  mount; it is the next piece of work, so that one directory holds everything.
- **The agent and the secrets** is an app change that comes with this design:
  in `run_mode = agent`, the app no longer requires the app secret or the JWT
  secret. Until images with it are released, the agent of a stack deployed by
  this installer does not start.
- **The app's own fixes this work found:** an empty CORS origin list makes it
  panic at boot, and it exits when the database is not reachable at boot
  instead of retrying. The installer works around both (§5, §2 step 8).

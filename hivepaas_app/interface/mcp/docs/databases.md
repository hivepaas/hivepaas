# Databases on HivePaaS

A database or a cache is an app like any other. What makes it one is its
**App Kind** (`get_app_settings`, kind `kind`): its category (`database`,
`cache`), its engine, and its credentials - database name, user, password, and
for MySQL and MariaDB a root password. HivePaaS publishes them to the app
itself as `HIVEPAAS_DATABASE_NAME`, `HIVEPAAS_USER`, `HIVEPAAS_PASSWORD`,
`HIVEPAAS_ROOT_PASSWORD`, and shares all but the root password with the other
apps of its env. The credentials live there and nowhere else.

## Connect an app to a database

Apps of the same env reach a database on their private network, by its key and
port, such as `db:5432`. Never copy a password into an app: refer to the
database's shared variables, resolved on each deployment.

1. `list_env_link_targets` for the app: the databases, caches and stores it can
   link to.
2. `get_env_link_suggestions` with the target: groups of variables, the
   recommended one marked - a connection URL such as `DATABASE_URL`, the parts
   one by one, the engine clients' own (`PGHOST`, `PGSSLMODE`...). Each value is
   a reference, such as `${db.HIVEPAAS_PASSWORD}`; keep it as it is.
3. `plan_update_app_settings`, kind `env-vars`: add them to `runtimeEnvVars`.
   The list is replaced whole: send the existing variables too.

## Run a database from its own image

An app from the app store is set up already. An app run from an engine's image
by hand - `postgres:18`, `mysql:8.4` - reads its first user and password from
env vars:

1. Its App Kind: category `database` (or `cache`), its engine, its credentials.
2. `get_env_self_suggestions`: the variables the engine's official image reads,
   such as `POSTGRES_PASSWORD=${HIVEPAAS_PASSWORD}`, each a reference to the
   App Kind. Its warnings say what App Kind lacks.
3. Add them with `plan_update_app_settings`, kind `env-vars`.

PostgreSQL, MySQL, MariaDB, MongoDB, ClickHouse and RabbitMQ read them only
when they first create their data (`initOnly`): on a database with data, a
password is changed inside it, not here. Redis and Valkey read none: the
suggestion is a command, set as the deployment settings' `command`.

## Reach a database from outside the cluster

Only when something outside must, such as a desktop client. It is the app's
**routing** settings (`get_app_settings`, kind `routing`), never its App Kind:

- `exposePublicly: true`, and a domain with `protocol: tcp`, its
  `containerPort` the database's port, and `enabled: true`. Traefik opens that
  port on every node and routes by the domain the client's TLS handshake names
  (SNI); the client must connect with TLS.
- Traefik ends TLS with the domain's certificate, unless `tlsPassthrough` is
  on: then the database serves its own certificate, mounted into it.
- PostgreSQL, MongoDB, Redis and Valkey work this way. **MySQL and MariaDB do
  not** - the server speaks first, before any TLS: publish their port in the
  network settings instead.
- A client that offers an ALPN protocol HivePaaS does not list is refused with
  `no application protocol`: add it to the domain's `extraAlpnProtocols`.
- The person opens the port in their firewall; say so.

## SSL mode

App Kind's `sslMode` is what the apps of the env connect with - straight to the
database, never through Traefik. A database with no TLS of its own, the usual
case, takes `disable`; one whose own TLS is on, `require`. A client outside
chooses its own: `verify-full` with the system's authorities
(`sslrootcert=system` for libpq).

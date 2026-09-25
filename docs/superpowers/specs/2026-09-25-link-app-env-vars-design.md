# Linking an App's Variables to Another App

An app reaches another app of its env through variables: `${postgres.HIVEPAAS_HOST}`
names what the app `postgres` shares. Writing those by hand means knowing the
target's key, which variables its kind shares, and how a client of that engine
wants them put together. This design has HivePaaS suggest them.

On an app's Env Variables screen, a **Link App** button opens a dialog. The
person picks an app of the same env; HivePaaS suggests variables for it, in
groups - a connection string, the parts one by one, the target's own shared
variables - and the chosen ones are added to the form, to be saved as usual.

---

## Decisions

1. **The backend suggests.** What a target shares depends on its kind, its
   engine and its own shared variables, which the backend knows. The dashboard
   shows and picks.
2. **Suggestions are references, never values.** Every suggested value is a
   `${<app>.<VAR>}` expression, so the endpoint reveals nothing and needs no
   Reveal Secrets permission. The value is resolved when the app is built, as any
   reference is.
3. **Nothing is saved by the dialog.** It adds rows to the form; the person
   reviews them, and Save saves them. "Show Final Values" works on them at once.
4. **A password in a URL is URL-encoded.** A new shared variable,
   `HIVEPAAS_PASSWORD_URLENCODED`, holds the password percent-encoded, and every
   connection string uses it. A password with `@`, `:` or `/` would otherwise
   break the URL it is put in, silently.
5. **Suggestions come from a table of recipes in Go**, by category and engine.
   An engine the table does not know gets its category's recipe.

## 1. The new system variable

`HIVEPAAS_PASSWORD_URLENCODED` is published by apps of kind `database` and
`cache`, next to `HIVEPAAS_PASSWORD`, and shared the same way:

- **Value.** The password with every byte outside RFC 3986's unreserved set
  (`A-Z a-z 0-9 - . _ ~`) percent-encoded. That is safe in a URL's user info,
  path and query alike. An empty password gives an empty value.
- **Where.** `kindEnvVars` publishes it; `base.AppKindSharedEnvVars` lists it
  for `database` and `cache`, which the test that holds the two together
  checks; `mapAppUnallowedVar` reserves the name, so a person cannot define it.
- **Not added:** an encoded user. HivePaaS's own users are plain identifiers,
  and a person who picks an unusual one can edit the suggested string.

## 2. The API

Both under `/projects/{projectID}/{projectEnv}/apps/{appID}/env-vars/`, read
permission on the app.

**`GET link-targets`** - the apps the dialog offers:
- the apps of the same env, without the app itself and without preview apps;
- each as `{id, key, name, category, engine}`, `category` and `engine` empty
  for an app with no kind;
- ordered by name.

**`GET link-suggestions?targetAppId=<id>`** - what to add for one target:

```json
{
  "data": {
    "target": {"id": "...", "key": "postgres", "name": "Postgres", "category": "database", "engine": "postgres"},
    "groups": [
      {
        "id": "connection-url",
        "title": "Connection string",
        "description": "One URL most client libraries accept.",
        "recommended": true,
        "warnings": [],
        "vars": [
          {"key": "DATABASE_URL",
           "value": "postgres://${postgres.HIVEPAAS_USER}:${postgres.HIVEPAAS_PASSWORD_URLENCODED}@${postgres.HIVEPAAS_HOST}:${postgres.HIVEPAAS_PORT}/${postgres.HIVEPAAS_DATABASE_NAME}?sslmode=${postgres.HIVEPAAS_SSL_MODE}",
           "description": "Connection URL"}
        ]
      }
    ]
  }
}
```

- **Refused:** a target that is not in the app's env, the app itself, or a
  preview app - `ERR_APP_NOT_FOUND`.
- **Warnings** are strings on a group, shown above its rows:
  - the target has no container port: "Postgres has no container port; set one
    in its Routing Settings, or the port below is empty";
  - the target has no kind: only the common group and its shared variables are
    offered, and the dialog says why.
- **Keys** are suggestions: the dialog lets the person change each one and adds
  a prefix.

## 3. The recipes

A recipe is a list of groups for a category and a set of engine names. Engine
names are matched lowercased, after aliases:

| engine family | aliases |
|---|---|
| postgres | postgres, postgresql, timescaledb, postgis, pgvector |
| mysql | mysql, mariadb, percona |
| mongodb | mongodb, mongo, ferretdb |
| redis | redis, valkey, keydb, dragonfly |
| memcached | memcached |

`A` below stands for the target's key. Every group's vars are listed in order.

**database / postgres**
1. *Connection string* (recommended) - `DATABASE_URL=postgres://${A.HIVEPAAS_USER}:${A.HIVEPAAS_PASSWORD_URLENCODED}@${A.HIVEPAAS_HOST}:${A.HIVEPAAS_PORT}/${A.HIVEPAAS_DATABASE_NAME}?sslmode=${A.HIVEPAAS_SSL_MODE}`.
2. *Individual variables* - `DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSL_MODE`.
3. *libpq variables* - `PGHOST, PGPORT, PGUSER, PGPASSWORD, PGDATABASE, PGSSLMODE`: what libpq and most Postgres drivers read with no configuration.

**database / mysql**
1. *Connection string* (recommended) - `DATABASE_URL=mysql://${A.HIVEPAAS_USER}:${A.HIVEPAAS_PASSWORD_URLENCODED}@${A.HIVEPAAS_HOST}:${A.HIVEPAAS_PORT}/${A.HIVEPAAS_DATABASE_NAME}`.
2. *Individual variables* - `DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME`.

**database / mongodb**
1. *Connection string* (recommended) - `MONGODB_URI=mongodb://${A.HIVEPAAS_USER}:${A.HIVEPAAS_PASSWORD_URLENCODED}@${A.HIVEPAAS_HOST}:${A.HIVEPAAS_PORT}/${A.HIVEPAAS_DATABASE_NAME}?authSource=admin`.
2. *Individual variables* - `MONGO_HOST, MONGO_PORT, MONGO_USER, MONGO_PASSWORD, MONGO_DATABASE`.

**database / any other engine** - *Individual variables* (recommended): `DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME`.

**cache / redis**
1. *Connection string* (recommended) - `REDIS_URL=redis://:${A.HIVEPAAS_PASSWORD_URLENCODED}@${A.HIVEPAAS_HOST}:${A.HIVEPAAS_PORT}/0`.
2. *Individual variables* - `REDIS_HOST, REDIS_PORT, REDIS_PASSWORD`.

**cache / memcached** - *Servers* (recommended): `MEMCACHED_SERVERS=${A.HIVEPAAS_HOST}:${A.HIVEPAAS_PORT}`.

**cache / any other engine** - *Individual variables* (recommended): `CACHE_HOST, CACHE_PORT, CACHE_PASSWORD`.

**storage / any engine** - an S3 API:
1. *S3 client* (recommended) - `S3_ENDPOINT=http://${A.HIVEPAAS_HOST}:${A.HIVEPAAS_PORT}`, `AWS_ACCESS_KEY_ID=${A.HIVEPAAS_KEY_ID}`, `AWS_SECRET_ACCESS_KEY=${A.HIVEPAAS_SECRET}`, `S3_BUCKET=${A.HIVEPAAS_BUCKET}`, `AWS_REGION=${A.HIVEPAAS_REGION}`.

**webapp, or no kind**
1. *Addresses* (recommended) - `<K>_URL=http://${A.HIVEPAAS_HOST}:${A.HIVEPAAS_PORT}` (inside the cluster) and `<K>_PUBLIC_URL=${A.HIVEPAAS_APP_URL}` (from outside), where `<K>` is the target's key uppercased, hyphens made underscores.

**Every target, last** - *Shared variables of <name>*: each variable the target
declares shared, as `<KEY>=${A.<KEY>}`. Omitted when there are none. Its system
variables are not repeated here; the groups above cover them.

## 4. The dialog

In `EnvVarsBaseForm`, a **Link App** button sits before "Show Final Values" on
the app's screen only - not the project's or env's, which have no apps to link.
It opens a dialog 1000px wide:

- **Target.** A combobox of `link-targets`: name, key, and a badge of category
  and engine. Choosing one loads its suggestions.
- **Options.** A *prefix* input, empty by default, applied to every key - for an
  app that links two databases. A *section* switch: Runtime (default) or
  Build-time.
- **Groups.** One card per group: title, description, warnings, a checkbox that
  selects the whole group, and a row per variable - a checkbox, the key as an
  editable input, the value read-only in monospace. Recommended groups start
  selected.
- **Conflicts.** A key the form already has, in the chosen section, is marked
  "exists"; its row offers *replace* or leaves the person to rename it. Two
  selected rows with one key are refused until one is renamed.
- **Add N variables** appends the selected rows to the chosen section of the form
  and closes the dialog; a replaced key takes the new value in place. Nothing is
  saved.

## 5. Testing

- **Encoding.** `HIVEPAAS_PASSWORD_URLENCODED` of `p@ss:w/rd?#%` and of an empty
  password; the shared-variable list and the publisher agree.
- **Recipes.** Each engine family and each category fallback gives its groups;
  aliases resolve; every `${A.VAR}` a recipe writes is a variable the target's
  category shares (checked against `AppKindSharedEnvVars` and the common list).
- **API.** Targets exclude the app and previews; suggestions refuse a target in
  another env; a target with no port carries the warning; a target's shared
  variables form the last group.
- **Dashboard.** A run against the backend: link a Postgres app, add the URL,
  save, and "Show Final Values" shows it resolved.

## Not in this design

- **Apps of other envs or projects.** A reference resolves within one env.
- **Checking that the two apps share a Docker network.** Apps of one env share
  its network unless someone detached one.
- **Recipes a template declares.** A template could name the variables its app
  wants; the table in Go comes first.

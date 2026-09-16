# Configuration Spec

HivePaaS has no way to write down what a system is configured to be. Everything
lives in the database and in the Docker Swarm services it manages, and the only
way to reproduce a project on another installation is to click through it again.

This adds a **spec**: a declarative, human-readable snapshot of configuration at
a chosen scope, exported as a portable bundle, designed so that it can later be
imported to provision the same configuration on a different installation.

Export is built now. Import is not - but its contract is specified here, because
a format designed without knowing what will read it is a format that will be
wrong.

## Scope

**In:** export at global, project, project-env and app scope; settings plus the
configuration that lives in the Swarm service spec; references between objects;
three secret handling modes; a documented import contract.

The common case it is built for is **re-import into the same installation** -
restore, rebuild, clone. Migration to another installation is supported through
the `skip-missing` import flag rather than by weakening what export records.

**Out, deliberately:** the import implementation; volume *data* (the spec
describes a volume, it does not carry its contents); the database itself (that
is `sysbackupservice`, a different tool for a different job); users, roles and
ACLs; audit history, tasks, deployments and any other record of what happened
rather than what is configured.

## The rule that shapes everything: scope contains its children

Exporting a scope exports that scope's own data and everything below it.
Nothing above it, nothing beside it.

```
global      -> global + hivepaas scope settings, then every project
project     -> the project, its envs, and their apps
project-env -> the env and its apps
app         -> the app
```

**Child apps are not exported.** The only thing that sets `App.ParentID` is
`apppreviewservice` - "Preview app must be a child app of the current" - so a
child app is always a preview environment for a pull request or branch, created
when the PR opens and destroyed when it closes. The same holds for the database
apps cloned alongside one, linked as `ResourceTypeLogicalChildApp`.

They are derived state with a lifecycle of their own, like `backup-snapshot`:
restoring a preview of a PR that closed months ago is noise, and the preview
machinery would remove it anyway. Export skips them and the report says how
many.

A reference that points outside the exported scope cannot be followed, and is
written as an **external reference** instead (see section 3). This is the single
most common thing an import will have to cope with, and it falls directly out of
this rule rather than being an edge case bolted on later.

---

## 1. Bundle layout

One export is always one `.tar.gz`, whatever the scope. A single-project export
is a bundle with one project directory in it. One artifact, one code path, one
place for the manifest to live.

```
hivepaas-spec-20260916T101500Z.tar.gz
├── spec.yaml                          manifest
├── global.yaml                        global + hivepaas scope settings
└── projects/
    ├── project_a/
    │   ├── project.yaml               project-scope settings
    │   └── envs/
    │       ├── dev.yaml               env settings + every app in dev
    │       └── prod.yaml
    └── project_b/
        ├── project.yaml
        └── envs/
            └── dev.yaml
```

Both path segments are safe by construction:
`idx_uq_projects_key ON projects(key) WHERE deleted_at IS NULL` makes a project
key unique across the installation, and
`idx_uq_project_envs_key ON project_envs(project_id, key)` makes an env key
unique within its project.

### Why the file boundary is the env

Partial import - selecting part of a bundle rather than all of it - does not
require this. A selector is a path filter applied after parsing, and picking one
subtree out of a 9,000-line document is trivial. The layout is an ergonomics
decision, not a capability one.

It is decided by which boundaries people actually select at:

| selection | per project | **per env** | per app |
|---|---|---|---|
| global only | file | file | file |
| global + chosen projects | file | file | file |
| one project's own settings | filter | **file** | file |
| one project + one env | filter | **file** | file |
| one app | filter | filter within one env file | file |
| files for 10 projects x 3 envs x 10 apps | ~12 | **~42** | ~300 |
| typical file size | 6-9k lines | ~1.5k lines | ~150 lines |

Per env puts a file boundary under four of the five selections, which makes
partial import a subset of the manifest's `files:` list - implementable as a
tree of checkboxes over exactly that list. The fifth is inherently surgical and
parsing one env file to reach one app is not a burden.

An env file is also the unit people review. "What differs between staging and
production" is a diff somebody genuinely wants; a single app in isolation
usually is not. Going all the way to a file per app costs 300 files and buys
only that fifth row.

`filearchiver.CompressTarGz` / `DecompressTarGz` already exist and are tested;
the exporter writes a temp directory and archives it.

### The manifest

```yaml
apiVersion: hivepaas.com/v1
kind: Spec
exportedAt: 2026-09-16T10:15:00Z
sourceAppVersion: v0.1.0      # base.StableVersion.AppVersion - informational
sourceVersionCode: v000001    # base.CurrentVersion - what compatibility is judged on
scope: global
secretsMode: none          # none | encrypted | plaintext
files:
  - global.yaml
  - projects/project_a/project.yaml
  - projects/project_a/envs/dev.yaml
  - projects/project_a/envs/prod.yaml
```

Every payload file also carries a minimal header of its own (`apiVersion`,
`kind`, `scope`, and the project key where applicable), so that a file extracted
from the bundle and handed to somebody on its own still says what it is.

### Determinism

Two exports of an unchanged system must produce byte-identical YAML, or the
format is useless for review and for git. Everything is ordered explicitly:

- projects by `key`
- envs by `Index`, then `key`
- apps by `key`
- settings by type, then by the key derived in section 3
- map-valued fields (labels, driver options, sysctls) by key

Ordering is by the field named, never by database return order.

---

## 2. What is exported

### There are two sources of truth, not one

The obvious model - "configuration is the `settings` table" - is wrong, and
building on it produces a spec that cannot provision an app.

Splitting the app configuration usecases by where they read from:

```
From the Swarm service                 From the settings table
  deployment_settings_get.go             env_vars_get.go
  resource_settings_get.go               routing_settings_get.go
  storage_settings_get.go                kind_settings_get.go
  network_settings_get.go
  container_settings_get.go
  service_settings_get.go
```

Every file on the left calls `clusterService.ServiceInspect(ctx, app.ServiceID)`.
Volume mounts, networks, CPU and memory limits, replicas, restart policy,
security options, capabilities, sysctls and ulimits appear nowhere in
`hivepaas_app/entity/` for an app - they live in the Swarm service spec.

A spec built only from `settings` yields an app with the right environment
variables, the right domains and the right engine, and no storage, no networks,
no resource limits and default replicas.

### Read `swarm.Service` directly, not the DTOs

The exporter reads the Swarm service itself, through
`clusterService.ServiceInspect`, and maps it onto spec types it owns. It does
**not** call `appsettingsdto.Transform*()`.

The reason is layering. `ARCHITECTURE.md` places `dto` above `service`, and a
DTO is the wire shape for the dashboard: it is expected to change when the UI
changes. A service reaching up into `usecase/<x>uc/<x>dto` inverts the
dependency, and it couples a **durable, versioned artifact** to a contract that
is free to move for reasons that have nothing to do with the spec. A UI redesign
would silently change the spec format.

The rule is held almost everywhere in the codebase today - one service imports a
usecase DTO (`taskservice` -> `taskdto`), and the others reach only for
`usecaseagent/*dto`, which are agent RPC wire types rather than UI shapes. This
design does not add a large new exception for five blocks at once.

The cost is field mapping written twice, and three pieces of distillation the
Transform functions were doing for free. The spec layer now owns all three:

- `Networks[].Target` is a network **id** (`8vo4p3pwm1aksdu2ilryn8mpf`), not a
  name. `specservice` resolves it via `networkService` - a service depending on
  a service, which the layering allows.
- Durations come back as nanosecond integers and serialize as `3e+10`. The spec
  types use `timeutil.Duration` and convert on the way out.
- Raw Swarm carries fields that have no place in a declarative document -
  `Placement.Platforms` (runtime-detected, `[{Architecture: amd64}]`),
  `ForceUpdate`, `Runtime`, `Version.Index`, `CreatedAt`, `UpdatedAt`.

That last one turns from a cost into a benefit. The spec now has an **explicit
allowlist** of exported fields rather than inheriting whatever a DTO happens to
expose, so a field added to Swarm or to HivePaaS cannot appear in a spec because
nobody noticed.

### Keeping the round-trip guarantee without the coupling

Owning the field list gives up the one thing DTO reuse bought for free: the Get
and Update DTOs carry identical field sets, so anything readable was writable.

That guarantee is recovered in tests rather than at runtime. A test may import
anything - a test is not a layer - so `specservice`'s test suite uses the Update
request DTOs as an **oracle**:

- every field the spec exports must correspond to a field some `Update*Req`
  accepts, or export is producing something import can never write back
- every field `Update*Req` accepts and the spec does **not** export is reported
  as a coverage gap

The second half is what replaces "a field added to a DTO propagates
automatically": instead of appearing in the spec silently, a newly configurable
field makes a test fail with a message naming it. For a format that has to stay
stable across versions, being told is better than being changed.

### The `deployment` block

Swarm-derived configuration is grouped under `deployment`, beside the
`app-deployment` setting it belongs with, in named sub-blocks:

```yaml
deployment:
  source:     {...}   # app-deployment setting
  container:  {...}   # ContainerSpec
  resources:  {...}   # Resources
  storage:    {...}   # Mounts
  networks:   {...}   # Networks + EndpointSpec
  service:    {...}   # Mode + Placement
```

Sub-blocks rather than one flat block, for two reasons. `app-deployment` is a
real setting type with `Version` and `Migrate()`; the Swarm-derived part has
neither, and a future change to a Swarm field needs somewhere to hang a
migration that is not the setting's version. And `Command` and `WorkingDir`
exist in both. `TransformDeploymentSettings` resolves this today by reading
them from the service and letting the setting overwrite them; the exporter
applies the same precedence itself, so `source` carries the merged value and
`container` does not repeat it. The precedence is part of the spec's contract
and is pinned by a test, not inherited from a DTO.

**The whole block is optional.** An app that has never been deployed has an
empty `ServiceID` and `ServiceInspect` will fail on it. This is not
hypothetical: two of the five user apps in the development database are in that
state. Such an app exports its settings alone, and imports back to the same
state.

### What export is for, and the rule that follows

The common case is not migration. It is **re-import into the same installation**
- restore after a mistake, rebuild after losing the database, clone a project
beside itself. Migration to a different installation is the rarer case, and it
is handled at import time by a `skip-missing` flag, not by weakening the export.

So export is a **faithful snapshot**, and the strip list is governed by one rule:

> Strip what is **regenerated** or **transient**. Keep what is merely
> **system-specific**.

System-specific values are the entire point: they are what makes a same-system
restore exact. Regenerated values cause drift - the spec says one thing, the
system recomputes another, and every diff is noise. Transient values cause
*actions*, which is worse than noise.

A value that is useless on a different installation is not a reason to drop it.
On the same installation it is exactly right, and on a different one it either
resolves or `skip-missing` removes it.

### Setting types

46 types. Only three are skipped.

#### Skipped (3)

| type | why |
|---|---|
| `api-key` | `SecretKey` is a `HashField` - a one-way hash, not an encrypted value. An imported hash authenticates nobody, on any installation. Keys must be re-issued |
| `backup-snapshot` | restic snapshot ids, times and sizes, discovered by scanning the repository. Not configuration at any point |
| `app` | declared in `base/setting.go` with no registered parser and no rows. A dead type; worth deleting separately |

#### The `cluster-*` types: scope separates configuration from discovery

`cluster-node`, `cluster-network` and `cluster-volume` each exist in two
distinct flavours, and the difference is not in the payload struct - it is in
the scope.

| | discovered by sync | authored by HivePaaS |
|---|---|---|
| scope | `global` | `project` / `project-env` |
| `Setting.ID` | the Docker object id | a ULID |
| created by | `networks_sync.go`, `volumes_sync.go`, `docker_node_sync.go` - all three hardcode `ObjectScopeGlobal` | `networkservice/project_env.go`, `volumeservice/project.go` |
| after a database loss | recreated by the next sync | **gone, unless the spec has it** |

**Export the non-global rows. Skip the global ones.**

The authored rows are real configuration: which network belongs to which
project, which volume is a project's default, its driver, its options, its node
pinning. Sync cannot recreate any of it, because sync only ever writes
`ObjectScopeGlobal`. Losing the database and re-syncing gives back a flat list
of Docker objects with every project association erased.

The discovered rows are noise that regenerates. In the development database ten
of the thirteen `cluster-network` rows are `bridge`, `host`, `none`, `ingress`,
`docker_gwbridge` and Docker Desktop extension networks. Their `Setting.ID` is
the Docker object id, so on a different installation they are foreign
identifiers, and `SyncNetworks` soft-deletes any row with no matching Docker
object on its very next run. Exporting them costs noise and buys nothing.

This rule also settles `cluster-node` without a special case: no code anywhere
creates a node setting outside `docker_node_sync.go`, so there are no non-global
rows and nothing is exported. That is the right answer for it -
`ClusterNode{}` is empty, `Kind` comes from `node.Spec.Role`, and `Name` comes
from `node.Spec.Name` (verified: the name `123456` in the development database
is literally Docker's `Spec.Name`). Node labels live on the Docker node, not on
the setting. There is no state on a node setting that only the export could
restore.

#### `ssl-cert` is exported whole

No field is stripped - not the material, not the issuance state.

The certificate, private key and CA certificate are kept because on the same
installation they are exactly right, and on a different installation serving the
same domain they work immediately with no ACME round trip at all. If the target
serves a different domain the certificate is simply unused, which costs nothing.

Dropping `ExpireAt`, `RenewableFrom` and `NotifyFrom` would be an outright bug.
They are what the renewal scheduler reads; a restored certificate without them
is one the system does not know when to renew.

There is also a rate limit to respect. Forcing re-issuance on every restore
walks into Let's Encrypt's duplicate-certificate limit, which is five per week -
easily reached by anyone testing a restore.

The private key is an `EncryptedField`, so it already follows `secretsMode` and
is absent entirely from a `none` export. The sensitive half is gated by a
mechanism that already exists; it does not need a second one.

#### Stripped fields (5 rules, and only 5)

Everything here is regenerated or transient. Nothing here is merely
system-specific.

| what | class | why |
|---|---|---|
| `app-routing.Reset` | transient | a command flag. Importing `Reset: true` performs a reset |
| `logging.Backend.Ingest` / `.Query` when `Managed` | regenerated | the code says so: "derived from the deployed service when Managed, and ignored" |
| `Mount.Key` | regenerated | `sha256("type:%v:src:%v:target:%v")`, recomputed on read. Replaced in the spec by `target` - see below |
| Swarm runtime fields - `Placement.Platforms`, `ForceUpdate`, `Runtime`, `Version.Index`, `CreatedAt`, `UpdatedAt` | regenerated | detected by Docker, never declared |
| service and container labels written by HivePaaS, Docker or Traefik | regenerated | see below |

**Reversed from an earlier draft, by the rule above.** These were on the strip
list and are now kept, because each is system-specific rather than regenerated:

| what | why keeping it is correct |
|---|---|
| `Secret.SwarmRef`, `ConfigFile.SwarmRef` | the Swarm secret with that id still exists on the same installation |
| `env-var` entries with `IsSystem: true` | **this one is a correctness bug, not a preference.** `HIVEPAAS_ROOT_PASSWORD` is generated once, and the database volume was initialised with it. Regenerating it on restore leaves the application holding a password the database does not accept |
| the image digest (`@sha256:…`) | it pins exactly the image that ran, which is what a restore wants. Measured across the cluster, all nine `hivepaas_*` stack services carry one and the HivePaaS-created services do not, because HivePaaS controls `options.QueryRegistry` - so user apps rarely have a digest at all, and when they do it is worth keeping |
| `cluster-volume.NodeID` | the Docker node id is valid on the same installation |
| `ssl-cert` material and state | see above |
| `app-clone` settings | the saved clone configuration is worth restoring after a database loss like anything else |

### Mounts are keyed by `target`, not by position

`Mount.Key` is stripped, so a mount needs some other identity in the spec. It is
not the list index.

A positional identity is fine while the spec only carries configuration - a
reordered mount list re-imports to the same configuration either way. It stops
being fine the moment anything outside the spec points at a mount, which is
exactly what a future snapshot does when it pins volume data to one. Data
pinned to `mounts[0]`, a reorder, and a restore is data attached to the wrong
volume, with no error and no warning.

So mounts are keyed by `target`, the path inside the container. Two mounts at
one target is meaningless, which makes it the natural key - but it is **not
currently validated as unique**: `storage_settings_update.go` has no such check.
The exporter therefore asserts it and reports a duplicate target as an error
rather than silently emitting two objects with the same key.

Targets are used verbatim, quoted by YAML like any other key:

```yaml
storage:
  mounts:
    "/var/lib/postgresql/data": { type: volume, source: 01M2574KP8…, … }
    "/etc/app/config":          { type: bind,   source: /srv/app/conf, … }
```

### Labels

Service and container labels are exported minus everything regenerated on apply:

| removed | |
|---|---|
| `hivepaas.*` | `app.info`, `app.id`, `project.id`, `projectEnv.id`, `app.prevServiceMode`, `app.placementConstraints` - all rewritten by HivePaaS when the service is applied |
| `com.docker.stack.*` | written by `docker stack deploy` |
| `desktop.docker.io/*` | injected by Docker Desktop, not by HivePaaS, and carrying absolute host paths - observed on a live service: `desktop.docker.io/mounts/0/Source: /Users/…/.appdata/hivepaas` |
| `traefik.*` **unless** the key contains `.x-custom-` | exactly the rule `updateSwarmServiceLabels` uses when cleaning up its own labels. `.x-custom-` entries are the user's and are kept |

These are stripped because they are regenerated, not because they are
system-specific. Keeping them would make every spec disagree with the system the
moment the service is applied.

The Docker Desktop case is why this is a maintained denylist with a test rather
than a one-line prefix check: third parties write labels onto services too, and
the next one will not be named `hivepaas`.

### HivePaaS-derived placement constraints

HivePaaS injects placement constraints of its own, derived from `app-placement`
settings and from the node pinning of mounted volumes, and it records which ones
it added in the `hivepaas.app.placementConstraints` label.
`placement_apply.go` already filters them back out to recover the user's
constraints, and the exporter reuses that filter.

This is a regeneration case, and the consequence of getting it wrong is
concrete: without the filter an import duplicates every constraint, and carries
a `node.id==…` that the placement service did not put there and will not remove.

### The `hivepaas` project is excluded

There is a project with key `hivepaas`, holding nine apps - `app`, `worker`,
`db`, `redis`, `redis_ui`, `traefik`, `adminer`, `agent`, `updater` - all with
live service ids. It is HivePaaS's own stack, modelled as a project.

Exporting it and importing it elsewhere would redeploy HivePaaS's database and
proxy over the ones `install.sh` just created.

The constant already exists: `base.HivepaasProjectKey`, and
`base.UnallowedProjectKeys` already forbids users from creating a project with
that key. `global.yaml` carries global- and hivepaas-*scope* settings; it does
not carry the hivepaas *project*.

### Size

Rendering the Swarm-derived block of three live services to YAML:

| service | lines |
|---|---|
| `hivepaas_db` | 57 |
| `hivepaas_app` | 62 |
| `hivepaas_traefik` | 100 |

The spec's own allowlist trims this further. With environment variables, routing,
secrets and config files, an app is roughly 100-180 lines.

With the per-env layout that puts a ten-app env at roughly 1,500 lines, which is
comfortable to read and to diff. A project of fifty apps spread over three envs
is three files of that size rather than one of 6-9k.

---

## 3. Identity and references

### Singletons need no key

Settings fall into two classes, and the codebase already draws the line -
`SettingRepo.GetSingle(scope, type)` and the whole `setting_*_unique.go` family
on one side, `List` on the other.

A **singleton** has at most one instance per scope, so `(scope, type)` is its
identity and the block name in the YAML is its key:

```yaml
envVars:   [...]
routing:   {...}
kind:      {...}
features:  {...}
placement: {...}
```

The data agrees. Querying every setting for a missing name returns exactly two
types - `env-var` and `app-routing` - and both are singletons. Every collection
type has a name.

### Collection keys: the name, verbatim

A key is derived from the name and **not slugified**. Running
`slugify.SlugifyAsKey` over the certificate names actually in the development
database:

```
"*.dev.hivepaas.com"  ->  "dev_hivepaas_com"
"dev.hivepaas.com"    ->  "dev_hivepaas_com"     collision
"*.localhost"         ->  "localhost"
"localhost"           ->  "localhost"            collision
```

Slugifying destroys the wildcard, which is the whole difference between two
certificates - and here it collapses five certificates into three collisions
that `kind` cannot break, because `*.dev.hivepaas.com` and `dev.hivepaas.com`
are both `self-signed`.

Keys are used inside YAML, not as filenames, and YAML quotes anything. So:

```
key = name
  duplicate within (scope, type)  ->  name@kind
  still duplicate                 ->  name@kind#2, #3 …
ordered by created_at ASC, then id ASC
```

Applied to the same five certificates, only the genuinely ambiguous pair gains a
suffix:

| name | kind | key |
|---|---|---|
| `*.dev.hivepaas.com` | self-signed | `*.dev.hivepaas.com` |
| `*.localhost` | letsencrypt | `*.localhost` |
| `dev.hivepaas.com` | self-signed | `dev.hivepaas.com` |
| `localhost` | letsencrypt | `localhost@letsencrypt` |
| `localhost` | self-signed | `localhost@self-signed` |

`kind` is empty for `ssh-key`, `notification` and `basic-auth`, so duplicates
there fall straight through to `#n`.

Each object also carries `id:`, its source ULID. It is not what references point
at - a key is - but it is what makes the common case exact. Re-importing into
the installation that produced the bundle matches on `id` and updates in place,
so an object that was renamed between export and import is still recognized as
the same object rather than created again beside itself. Key matching is the
fallback, used when the id is not found: on a different installation, always;
on the same one, only for objects created since.

For the `cluster-*` rows that are exported, `id` is a real ULID
(`volumeservice/project.go` and `networkservice/project_env.go` both generate
one), so this behaves like any other setting. The rows whose `Setting.ID` is a
Docker object id are the sync-discovered ones, and those are not exported.

### References are scope paths

```yaml
routing:
  domains:
    - domain: api.example.com
      sslCert: global/ssl-certs/localhost@self-signed
```

A path, not a bare name, so the scope is part of the reference and there is no
cross-scope ambiguity - which is what the original concern about duplicate names
on import was really about.

A reference pointing outside the exported scope becomes an external reference
carrying enough to find its target again:

```yaml
      sslCert:
        external: { type: ssl-cert, name: localhost, kind: self-signed, id: 01M2JQ… }
```

### Rewriting references on import

Reading references is solved: `GetRefObjectIDs()` is mandatory on the
`SettingData` interface and every type implements it correctly, because resource
link integrity already depends on it.

Writing them is not. There is no setter, and references live in four different
container shapes:

| shape | examples |
|---|---|
| `entity.ObjectID` | `SchedJob.App`, `SSLCert.Provider`, `BackupRepo.Volume`, `AppDomain.SSLCert` |
| `entity.ObjectValue` | `CommandTemplate.Script` |
| `entity.ObjectIDSlice` | `AppCloneSettings.CommandPipes` |
| a bespoke struct with a bare `ID string` | `SystemBackupCloudStorage`, `SchedJobCommandOutputFileStorage` |

The fourth shape is the trap. A reflection walk keyed on the *type* `ObjectID`
silently misses those two, and a missed reference imports as a pointer to an
identifier from another system - no error, no warning. This is the same class of
bug `reencryptValue` was written to avoid, where a textual pass would have
re-encrypted every `HashField` and destroyed every API key.

**Remap by value, not by type.** `GetRefObjectIDs()` already says exactly which
identifier strings are references. The shape they are stored in does not matter:

1. `refIDs := data.GetRefObjectIDs()` - authoritative, per type, already correct
2. walk the parsed value (the shape of `reencryptValue`) and replace any string
   field whose value is in that set
3. call `GetRefObjectIDs()` again and assert the result equals the expected new
   identifiers

Step 3 is what makes this safe. A missed reference fails loudly at import time
instead of becoming a dangling pointer, and a setting type added later is
covered without anybody remembering this document exists.

The only theoretical failure is a string that coincidentally equals a 26-character
ULID without being a reference. If a value equals a reference id, it is that
reference.

---

## 4. Secrets

### Three modes

| mode | capability | for |
|---|---|---|
| `none` *(default)* | none | structure only - review, git, handing to a colleague |
| `encrypted` | reveal + audit | a real backup |
| `plaintext` | reveal + audit + explicit confirmation | moving an installation, accepting the risk |

`none` is the default because it is the only mode that needs no capability and
leaks nothing.

### Reuse the existing seam

`usecase/settings/setting_secrets.go` already has what export needs:

```go
type secretDecrypter interface{ Decrypt() error }
```

Its comment states the intent: reaching decryption through one interface "is
what lets one place cover all of them - and what makes a setting type added
later covered without anybody remembering to come back here". Export calls
`revealSecrets()` and gets the capability check and the audit record for free,
with no type left out.

It also already enforces the scope boundary this spec needs:

```go
// An inherited setting is read through, not owned, by this scope
if setting.ObjectID != setting.CurrentObjectID { return nil }
```

A setting inherited from an outer scope is not decrypted. A project-scope export
cannot extract global secrets - the same rule that governs the whole format,
already implemented.

### Encryption does not use the app secret

`encrypted` mode asks the user for a passphrase. It does not use the system's
key-encryption key, for three reasons, of which the second decides it.

**The app secret protects every secret in the database.** Using it as the export
key means that to import, someone must type the source installation's app secret
into the target. The most sensitive credential in the system travels with the
file.

**It does not work.** Stored ciphertext is sealed with the *data* encryption key
- 32 random bytes generated per installation, kept wrapped by the app secret.
The target has a different DEK, so ciphertext from the source is meaningless
there. Making it work would mean shipping the wrapped DEK as well, at which
point the artifact is a clone of the installation's key material, not a spec.

**The codebase already answered this.** `sysbackupservice` faced the same
question and chose a separate user-supplied passphrase:

```go
encSecret, _ := data.SysBackupSettings.Encryption.Secret.GetPlain()
recipient, _ := age.NewScryptRecipient(encSecret)
encW, _ = age.Encrypt(w, recipient)
```

### The whole bundle is age-encrypted

`encrypted` mode produces `hivepaas-spec-….tar.gz.age`, using
`base.FileEncryptionFormatAge` and the same `age` path as system backup.

No hand-written cryptography, no per-value key derivation, and the file can be
decrypted with the standard `age` CLI without HivePaaS running - which matters
for an artifact whose purpose is to survive the system that produced it. age
carries its own salt in its header, so the bundle needs no key material of its
own.

The cost is that an encrypted bundle is opaque: the structure cannot be reviewed
or diffed without the passphrase. That is the correct trade for a backup
artifact, and `none` remains available for anything meant to live in a
repository.

`api-key` is skipped entirely, since a `HashField` cannot be decrypted by
anybody.

---

## 5. Import contract

Import is not built here. This is what the format promises it, and what it must
promise back.

### Two phases

**`validate`** parses the bundle, resolves every reference, checks every
collision, and **writes nothing**. It produces the full report.

**`apply`** writes, softening or skipping exactly what validate flagged, and
returns the same report with actual outcomes.

Validate is also a feature in its own right: preview before import.

### Selecting part of a bundle

A bundle is not all-or-nothing. The import request carries a selector, and
because the file boundary is the env, a selection is a subset of the manifest's
`files:` list plus an optional filter inside one of them.

```yaml
include: [global, projects/project_a]
exclude: [projects/project_a/envs/**]     # the project's own settings, no envs
```

Include plus exclude rather than a bespoke "this node only" syntax: the two
lists cover every selection without new grammar, and the pattern already exists
in the codebase as `AppCloneSettings.IncludedVolumes` / `ExcludedVolumes`.

The five selections this is built for:

| what the user wants | selector |
|---|---|
| global settings only | `include: [global]` |
| global plus some projects | `include: [global, projects/project_a, projects/project_c]` |
| one project's own settings | `include: [projects/project_a]`, `exclude: [projects/project_a/envs/**]` |
| one project plus one env | `include: [projects/project_a/project.yaml, projects/project_a/envs/dev]` |
| one app | `include: [projects/project_a/envs/dev/apps/backend]` |

### Selection creates a third kind of reference

Before partial import a reference was either inside the exported scope or
outside it. Selection adds a third case, and it needs its own report code even
though it reuses the same machinery:

| state | handling | report code |
|---|---|---|
| in the bundle, selected | rewritten to the new id | - |
| **in the bundle, not selected** | treated as external | `REF_NOT_SELECTED` |
| not in the bundle at all | treated as external | `REF_NOT_FOUND` |

The two codes behave identically - the reference is cleared, the object is
marked `pending` - but they call for opposite actions. `REF_NOT_FOUND` means the
user has to create something. `REF_NOT_SELECTED` means they only have to import
again with a wider selection. Reporting both as "not found" sends somebody off
to hand-build an object that is sitting in the file they just imported.

`--with-dependencies` expands a selection to the transitive closure of its
references within the bundle, computed from the same `GetRefObjectIDs()` the
remap uses. The closure follows **references**, never scope containment:
selecting one app pulls in the certificate it uses, and never its sibling apps.

### Import is lenient, and says what it did

A failed reference must not fail the import. The user gets the configuration
that could be applied, plus a list of what could not.

| severity | examples | behaviour |
|---|---|---|
| `fixable` | external reference not found; node label absent; domain held by another app | clear that part, continue, record it |
| `skipped` | object collides with a unique constraint and cannot be merged | drop that object only |
| `blocked` | stop, write nothing | |

### `skip-missing` is what makes one artifact serve both cases

Export is a faithful snapshot, so a bundle carries values that are meaningful
only on the installation that produced it - Swarm secret references, Docker node
ids, generated system environment variables, an image digest. Re-imported at
home, every one of them resolves and the restore is exact.

On a different installation they do not resolve, and that is what `skip-missing`
is for. With the flag set, an unresolvable system-specific value is treated as
`fixable`: dropped, recorded, and the object it belongs to marked `pending`.
Without it, the same situation is `blocked`, so nobody silently half-imports a
production system by accident.

This is the whole reason export does not have to guess which kind of import is
coming. It cannot know, so it does not try: it records everything, and the flag
at the far end decides what to do with what does not fit.

| value | same installation | different installation, `skip-missing` |
|---|---|---|
| `Secret.SwarmRef` | resolves | dropped; the secret is re-created on next apply |
| `cluster-volume.NodeID` | resolves | dropped; the volume becomes unpinned |
| `env-var` `IsSystem` entries | resolve, and must - the database was initialised with them | dropped; HivePaaS regenerates them |
| image digest | resolves | dropped back to the bare tag |
| `ssl-cert` material | resolves | kept if the domain matches, unused if it does not |

`blocked` is exactly three cases and no more:

- `apiVersion` not supported
- a `Setting.Version` newer than this installation's - `ErrDataVerNewerThanSystemVer`,
  a spec from a newer HivePaaS, which cannot be migrated backwards
- the bundle is corrupt, or the passphrase is wrong

### Incomplete objects must be visible, not just reported

Leniency has a real hazard: an app that looks fully configured but has no
certificate and no volume, with the only evidence in a report the user has
closed.

An object that took a `fixable` issue is imported with `status: pending` rather
than `active`. `base.SettingStatusPending` is declared, is already in
`AllSettingSettableStatuses`, and `IsActive()` returns false for it - so an
incomplete setting is not picked up by the runtime and nothing half-configured
goes live. It currently has no other use anywhere in the code, so this collides
with no existing meaning.

### Import must be idempotent

The repair loop is: import, read the report, create what was missing, **import
the same bundle again**. The second run fills in what was cleared and flips
`pending` to `active`.

That only works if import matches by key and updates in place instead of
creating duplicates. Idempotency is a requirement of import, not an optional
refinement.

### Defaults

`is_default` is exported. On import, if the target already has a default for the
same **`(scope, type, kind)`**, the incoming object becomes non-default.

The scope of that comparison is not `(scope, type)`. The system's invariant is
per-kind - `UpdateClearDefaultFlag(scope, type, kind, exceptID)` in
`setting_base.go` - and the data shows why it matters: `oauth` has one default
each for `github`, `gitlab`, `google` and `gitea`, and `access-token` the same.
Comparing without `kind` would demote three of four and leave GitHub, GitLab and
Gitea with no default at once.

### The report

```yaml
issues:
  - severity: fixable
    code: REF_NOT_SELECTED
    path: projects/project_a/envs/dev/apps/backend/routing/domains[0]/sslCert
    detail: { type: ssl-cert, name: localhost, kind: self-signed }
    availableIn: global.yaml
    action: cleared, setting marked pending
    hint: re-import including global.yaml
```

Readable enough to fix by hand, structured enough to drive a UI.

### Things import must not trust

- **`apps.global_key` can be stale.** An app in the development database carries
  `project_b_staging_frontend` while its `project_env_id` is `…:dev`. Paths are
  derived from `project_env_id`, never from `global_key`.
- **Domains and published ports are reservations**, recorded in `res_links` with
  `dst_type` `domain` and `port`. Import must check availability the way
  `appcloneservice` does, through `VerifyProjectDomains` and
  `VerifyDomainsAvailable`.
- **Seeded objects already exist.** `settinginitservice` creates default
  notification, domain, placement, image-build, cleanup, backup, renewal
  settings and four scheduled jobs on every fresh install. They are recognizable
  by `update_ver = 0` and their seeded names - not by `is_default`, which is
  false on all four jobs. These are matched and merged, never duplicated.

---

## 6. Extension points

Two, and deliberately no more.

**`apiVersion`** at the top of the manifest and every file. The one field that
makes every later decision possible.

**`options:`**, a free-form map on any object, which **export never writes and
import reads**. It is where an operator overrides what the spec cannot know -
the root domain for the target installation, a webhook URL, whether to follow
`RepoRef` instead of the pinned `CommitHash`.

Leaving it unwritten by export is the point: the format is open without anything
speculative being designed into it now.

---

## 7. Package structure

```
hivepaas_app/service/specservice/
  specmodel/              the spec types - the durable, versioned contract
                          Bundle, Manifest, Issue, Severity, and one type per
                          deployment sub-block (Container, Resources, Storage,
                          Networks, Service)
  specserviceimpl/        export: walk scopes, map Swarm -> specmodel,
                          apply per-type policy, serialize
  service.go              interface

hivepaas_app/entity/
  setting_spec.go         SpecPolicy per setting type, beside the parser registry
  setting_refmap.go       RemapRefs - the value-driven walk from section 3
```

`specmodel` depends on `entity`, `base` and the Docker types, and on nothing
above it. It is the whole point of owning the shape: the spec's contract is
pinned in one package that no UI change can reach.

Per-type policy is registered the way parsers already are:

```go
var _ = registerSettingParser(base.SettingTypeSSLCert, &sslCertParser{})
var _ = registerSpecPolicy(base.SettingTypeSSLCert, &sslCertSpecPolicy{})
```

A type without a registered policy is **skipped and reported**, never exported
blindly. A new setting type does not silently leak into the spec because
somebody forgot it; it shows up in the report as unclassified.

---

## What this needs from existing code

Reused as-is: `clusterService.ServiceInspect`, `networkService` for network id
resolution, `revealSecrets` and `secretDecrypter`, `filearchiver.CompressTarGz`,
`age` via `base.FileEncryptionFormatAge`, the placement-constraint filter in
`placement_apply.go`, `GetRefObjectIDs()` on every setting type.

New: the spec types themselves, the per-type spec policy registry, the Swarm ->
spec field mapping, the reference remap walk, the YAML serializer, the bundle
writer, and the export usecase plus handler.

Changed: nothing. Export is a read.

**A prerequisite for import, not for export.** The logic that applies these five
blocks back onto a Swarm service lives in `appsettingsuc`
(`applyAppResourceSettings` and its siblings), and `ARCHITECTURE.md` forbids one
usecase importing another. An import usecase therefore cannot call it. Before
import is built, that apply logic has to move down into a service both usecases
call - which is exactly what the layering rule prescribes: "When two usecases
need the same thing, that thing is a service." Export does not need this and
does not wait for it.

---

## Testing

- **Field-coverage oracle.** Since the exporter no longer shares types with the
  DTOs, a test compares the spec's allowlist against what the `Update*Req` DTOs
  accept, in both directions: a spec field with no writable counterpart is an
  error, and a writable field the spec omits is reported as a coverage gap. Test
  code may import the usecase layer; this keeps the round-trip property without
  a runtime dependency.
- **Round-trip per setting type.** Export a setting, write the result back, and
  assert the stored value is unchanged.
- **Reference remap self-check.** Table-driven across all four container shapes,
  including `SystemBackupCloudStorage` and `SchedJobCommandOutputFileStorage`
  explicitly. Assert the post-remap `GetRefObjectIDs()` matches, and assert that
  deliberately breaking the walk makes the test fail.
- **Determinism.** Export twice, assert byte equality.
- **Selection.** Each of the five selectors in section 5 against one fixture
  bundle, asserting exactly the intended objects are imported - in particular
  that `exclude: [projects/x/envs/**]` leaves the project's own settings and
  nothing else.
- **`REF_NOT_SELECTED` vs `REF_NOT_FOUND`.** Import an app whose certificate
  lives in `global.yaml` without selecting it, and assert the report says
  `REF_NOT_SELECTED` naming the file it is in. Then delete the certificate from
  the bundle and assert the same import reports `REF_NOT_FOUND`.
- **Dependency closure.** `--with-dependencies` on a single app pulls in the
  certificate it references and none of its sibling apps.
- **Mount targets are unique.** An app whose Swarm spec carries two mounts at the
  same target is reported as an error rather than exported with a duplicate key.
  Nothing validates this upstream, so the test is the only thing that holds it.
- **Preview apps are skipped.** A fixture with an app carrying a child app and a
  `LogicalChildApp` link, asserting neither is exported and the report counts
  them.
- **Label denylist.** A fixture carrying `hivepaas.*`, `com.docker.stack.*`,
  `desktop.docker.io/*`, `traefik.*` and `traefik.*.x-custom-*`, asserting
  exactly one survives.
- **Key derivation.** The five real certificate names, asserting the table in
  section 3.
- **`cluster-*` scope rule.** A fixture holding both flavours of each type -
  a sync-discovered row at global scope with a Docker id, and an authored row at
  project scope with a ULID - asserting only the authored ones are exported.
- **System environment variables survive a round trip.** Export an app whose
  `env-var` setting carries `IsSystem` entries, re-import into the same
  installation, and assert the values are byte-identical. This is the test that
  protects a restored application from a regenerated `HIVEPAAS_ROOT_PASSWORD`
  its database will not accept.
- **Unclassified type.** Register a setting type with no spec policy; assert it
  is reported, not exported.
- **Secrets.** Assert `none` emits no secret material; assert `encrypted`
  produces a file `age` can decrypt; assert an inherited setting's secrets are
  never decrypted.

---

## Extending to snapshots: spec plus data

A snapshot - configuration *and* the volume data behind it - is a wanted future
feature. It extends from this design, and the half this document does not cover
already exists in the repository:

| already built | |
|---|---|
| `services/backup/backupmodel` | an `Engine` interface with `Name() EngineType`; Kopia implemented |
| `backupreposervice` | repository init, sync, snapshot list and sync, cleanup, password change |
| `volumeservice.Rsync` | volume data copy - a direct host rsync fast path, and a container fallback running as a Swarm task |
| `appcloneservice.cloneVolumes` | **the consistency problem, already solved**: `stopSrcAppBeforeCloningVolumes` stops the source app for a clean copy unless `LiveVolumeClone` is set |

### It composes; it does not grow the bundle

Volume bytes do not go inside the `.tar.gz`. Three reasons, the second
decisive:

**Scale.** A spec is kilobytes, volume data is gigabytes. One artifact forces
streaming everywhere, and removes the ability to restore configuration alone
without moving the data with it.

**Kopia already does deduplication and incrementals.** Copying bytes into a
tarball throws all of it away: a daily snapshot of a 50 GB volume becomes 50 GB
a day instead of a delta.

**Two lifecycles.** Configuration changes when somebody edits it; data changes
continuously. Binding them into one artifact produces a duplicate copy of the
configuration on every data snapshot.

### `backup-snapshot` is the seam

This design **skips** `backup-snapshot` as history - "restic/kopia snapshot ids,
rediscovered by scanning the repository". In a snapshot bundle that same data
inverts its role and becomes the join between configuration and data:

```yaml
kind: Snapshot
apiVersion: hivepaas.com/v1
# the spec files, unchanged
data:
  repo: global/backup-repos/main
  volumes:
    - app: projects/project_a/envs/dev/apps/backend
      target: /var/lib/postgresql/data
      snapshotId: k4f9c1…
      takenAt: 2026-09-16T10:15:00Z
      sizeBytes: 12884901888
```

A mount is addressed as an `(app, target)` pair rather than by concatenating the
target onto the app path, since a target begins with `/` and concatenation would
produce an ambiguous double slash.

The bundle stays small. Restore is: import the spec, then `kopia restore` each
pinned snapshot into the volume the spec just recreated. `kind:` is already in
the manifest, so nothing about the format has to change to admit this.

### What this design already does for it

Keying mounts by `target` rather than by list position is the one thing done
now specifically so that a snapshot has something stable to pin data to. The
selector, the two-phase validate/apply, the severity ladder, `skip-missing` and
whole-bundle `age` encryption all carry over unchanged.

### Two things that do not carry over

**Idempotent re-import does not extend to data.** Re-importing a spec is safe by
design - match by key, update in place, run it as often as you like. Re-importing
volume data is a destructive overwrite. A snapshot restore must confirm rather
than inherit the spec import's "run it again" property.

**`secretsMode: none` becomes a misleading label.** Volume data holds secrets
whatever the mode says - a database volume contains its own passwords. A
snapshot carrying data has to force `encrypted` or `plaintext`, and refuse
`none`.

---

## Later, not now

- Import itself.
- A file per app, if the per-env file ever gets uncomfortable. The selector
  already addresses individual apps, so this would change file boundaries only.
- A `spec diff` against a live system.
- Re-importing a preview app. Excluded here because previews are ephemeral; if
  it is ever wanted, it belongs to `apppreviewservice`, not to the spec.
- Snapshots - see the section above, which is analysis rather than a plan.
- Exporting at user scope. Nothing there but `api-key`, which is skipped, so the
  scope is currently empty by construction.

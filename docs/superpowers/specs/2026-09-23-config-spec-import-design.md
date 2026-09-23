# Configuration Spec Import

Export (`2026-09-16-config-spec-export-design.md`) produces a bundle that nothing
can read back yet. This document designs reading it back: validating a bundle
against an installation, and applying it.

It implements the import contract in §5 of the export spec. Where the two
disagree, this document is the later decision and says why - see *Corrections to
the export spec*.

---

## What it is for

An operator needs configuration back in two situations: on the installation that
produced it, after something went wrong, and on another installation, when
moving. At each level of scope they want either all of it or part of it:

| level | all of it | part of it |
|---|---|---|
| global | the global configuration and every project | the global configuration and chosen projects |
| project | the project's configuration and every env | the project's configuration and chosen envs |
| env | the env's configuration and every app | the env's configuration and chosen apps |

## Decisions

1. **One selector, not six modes.** The six cases are include/exclude lists over
   one tree, and the dashboard shows that tree with checkboxes (§1).
2. **One code path for restore and migration.** There is no mode, and no
   detection of which installation a bundle came from. Objects are matched one at
   a time - id, then key, then created (§4) - and every value tied to an
   installation is checked against the target on its own (§6).
3. **The operator accepts what import will leave out, after seeing it.** Validate
   lists it; apply takes an explicit acceptance bound to that list (§3, §5).
4. **Import is a merge.** It creates and updates. It never deletes an object
   because the bundle does not have it.
5. **The API is stateless.** Every call carries the bundle, and the server keeps
   nothing between calls (§2).
6. **Three options**: what to do with objects that already exist, whether to
   deploy the apps import creates, and whether to deploy apps whose source
   changed (§3).
7. **Apply runs in three phases**: one database transaction, then Swarm updates
   for existing apps, then deployments (§8).
8. **One builder creates and updates apps.** The spec builder templates already
   use is extended to everything export writes, and runs on an existing service
   spec as well as a new one (§9).
9. **Validate is an export of the target, diffed against the bundle** (§5).
10. **Export starts writing source ids** (§4).

## Scope

**In:** validate and apply at global, project and project-env scope; selection
with dependency closure; restore and migration through one code path; the three
options; deploying after import; the dashboard flow; one change to export,
writing source ids.

**Out, deliberately:**

- **Cloning** a project, env or app under a new key. Id matching would find the
  original and overwrite it; a clone needs a step that remaps identity, which is
  a feature of its own.
- **Deleting.** An object present on the target and absent from the bundle is
  left alone, including an app created after the export. "Restore" here means
  "make what the bundle describes true", not "return to the state at export
  time".
- **The `hivepaas` project**, which export never writes.
- **Users, roles, ACLs, volume data and history**, which export does not carry.
  A project's owner is the exception: it travels as a reference to a user, and is
  resolved on import (§4).
- **Background import.** Apply runs in the request; §8 says why.
- **Import at app scope.** An app is imported at its env, by selecting it.

---

## Corrections to the export spec

1. **Export does not write ids.** Export spec §3 says every object carries `id:`,
   and its matching - id first, key as the fallback - depends on it. The code
   does not write them: `ProjectDoc`, `EnvDoc` and `AppDoc` have no id field, and
   `renderSetting` renders only a setting's data. The only source id in a bundle
   is on `ExternalRef`. Section 4 adds them. Nor did it write a setting's row:
   the export spec says `is_default` travels, and a collection's name and kind
   lived only in its entry key. Every setting body now carries its row under
   `setting` - name, kind, status, inheritable, default, version, expiry, ref id.
2. **The credential a restore must keep belongs to the app kind.** Export spec §5
   protects `env-var` entries flagged `IsSystem`, because a database volume was
   initialised with `HIVEPAAS_ROOT_PASSWORD`. That variable is not stored
   anywhere: `BuildSystemEnvVarsInApp` computes it from the `app-kind` setting
   whenever the app's environment is built, and the stored values are
   `AppKindDatabase.Password` and `RootPassword` - `EncryptedField`s, so secrets,
   cleared from an `omit` bundle. Section 7 is the rule for them, and the export
   spec's test *System environment variables survive a round trip* becomes
   *App-kind credentials survive a round trip*.
3. **There is no image digest to resolve.** `stripImageDigest` in `swarm_map.go`
   removes the `@sha256:` suffix at export, so the digest row of the export
   spec's `skip-missing` table has nothing to act on.
4. **`skip-missing` is an acceptance, not a mode.** The export spec's table reads
   as two modes, same installation and different installation. Each row is in
   fact decided value by value against the target (§6), and what is left for the
   operator is to accept what the plan will drop. The flag becomes `acceptIssues`
   on apply, bound to the plan it accepts by `planHash`.
5. **The apply-logic prerequisite is replaced.** The export spec requires moving
   the `apply*` functions of `appsettingsuc` into a service before import can be
   built. Import extends the spec builder instead, and runs it on an existing
   service (§9).
6. **A fixable issue does not always make its object pending.** The export spec
   marks every object that took a fixable issue `pending`. That is right when a
   reference was cleared, because the object is incomplete. It is wrong when the
   object is complete but different - a volume that lost its node pin, a routing
   setting that lost one of its domains - because `pending` stops an app that
   would have run. Section 6 gives the status per issue.
7. **A reference that left the export was written as a raw id.** The external
   form was designed but never wired: nothing registered the settings a
   reference reached outside the export. Export now loads them first and writes
   each such reference as `{external: {type, name, kind, id}}`. A reference to an
   app, a project, an env or a user is still the source id; import resolves it
   through the ids the bundle carries (§4), and on another installation one the
   bundle does not carry is `REF_NOT_FOUND`.

---

## 1. Selection

A selection is two lists of node paths, `include` and `exclude`. A path names a
node and everything below it, `*` matches one segment, and exclude wins. An empty
`include` means everything.

```
global                              global settings
projects/<p>                        a project and everything in it
projects/<p>/settings               the project's own record and settings
projects/<p>/envs/<e>               an env and every app in it
projects/<p>/envs/<e>/settings      the env's own record and settings
projects/<p>/envs/<e>/apps/<a>      one app
```

The six cases:

| case | selector |
|---|---|
| global, every project | `include: []` |
| global, chosen projects | `include: [global, projects/a, projects/c]` |
| project, every env | `include: [projects/a]` |
| project, chosen envs | `include: [projects/a/settings, projects/a/envs/dev]` |
| env, every app | `include: [projects/a/envs/dev]` |
| env, chosen apps | `include: [projects/a/envs/dev/settings, projects/a/envs/dev/apps/backend]` |

This refines the export spec's selector, which named files
(`projects/project_a/project.yaml`). An env's own settings share a file with its
apps, so a file path cannot select one without the other; the `settings` nodes
can.

Selecting a node whose ancestor the target lacks - an app whose env does not
exist - creates the ancestor's record, not its settings, within what the route's
scope allows (§3).

## 2. Flow and transport

The dashboard sends the bundle with every call, base64-encoded inside the JSON
body.

1. The operator picks a file, and the dashboard calls validate with an empty
   selection. An encrypted bundle sent without a passphrase fails with
   `ERR_SPEC_PASSPHRASE_REQUIRED`; the dashboard asks for it and calls again. The
   whole bundle is inside the age envelope, manifest included, so nothing can be
   shown before this.
2. The plan screen shows what validate returned (§5).
3. A change to the selection or to an option calls validate again.
4. Confirming calls apply with the `planHash` of the last validate.

Why the bundle is not uploaded first, through `/files/upload`:

- **The bundle is small.** Export measured an app at 100-180 lines of YAML.
  Compressed, an ordinary project is tens of kilobytes and a very large
  installation a few megabytes, so sending it again on each validate costs
  nothing that matters.
- **The upload API does not fit.** It refuses global scope. A `tmp` file lands
  under the `files` data path, which the system cleanup never sweeps - it removes
  dated directories under the temp dirs only. The file goes to whichever storage
  is configured, possibly S3, with a database row. A `plaintext` bundle would sit
  there with its secrets, and nothing would remove it when an operator abandons
  an import half way.
- **Uploading first does not give the guarantee that matters.** The target can
  change between validate and apply - a node removed, an app created, a domain
  taken - so apply has to validate again whatever the transport. What must be
  pinned is the plan, not the file, and `planHash` does that (§5).
- **It is how the API already takes content.** Project and app photos and config
  files arrive as base64 in JSON.

The import handlers wrap the body in `http.MaxBytesReader` at 25 MB, the limit
the git webhook handler uses (`webhookMaxBodySize`), and answer
`ERR_REQUEST_TOO_BIG` above it. The passphrase travels with every call, and is
never stored and never logged.

## 3. API

```
POST /spec/import/validate
POST /spec/import/apply
POST /projects/:projectID/spec/import/validate
POST /projects/:projectID/spec/import/apply
POST /projects/:projectID/:projectEnv/spec/import/validate
POST /projects/:projectID/:projectEnv/spec/import/apply
```

They are registered beside the export routes, in `router_spec.go`,
`router_projects.go` and `router_project_env.go`.

### Request

```json
{
  "bundle": "<base64 of the archive, as export produced it>",
  "passphrase": "<only for an encrypted bundle>",
  "selection": { "include": [], "exclude": [] },
  "options": {
    "existing": "update",
    "deployCreated": true,
    "deployChangedSource": false
  }
}
```

Apply takes the same body, plus:

```json
{ "planHash": "<from the last validate>", "acceptIssues": true }
```

| option | values | default | meaning |
|---|---|---|---|
| `existing` | `update`, `keep` | `update` | `update` makes an existing object match the bundle, and a running app whose configuration changes restarts. `keep` leaves every existing object alone: only what is missing is created, and nothing restarts |
| `deployCreated` | bool | `true` | queue the first deployment of every app import creates whose bundle entry has a `deployment.source`. Without it a created app runs the placeholder image, as an app created by hand does |
| `deployChangedSource` | bool | `false` | queue a deployment for every updated app whose `deployment.source` differs from the target's. Without it the app runs its new configuration on its current image until somebody deploys it |

`deployChangedSource` exists because an update cannot change the image. The rest
of an app's configuration reaches its running service at once (§8), but the image
changes only through a deployment.

### Validate response

```json
{
  "bundle": {
    "apiVersion": "v1",
    "scope": "global",
    "exportedAt": "2026-09-20T10:15:00Z",
    "sourceAppVersion": "v0.1.0",
    "secretsMode": "encrypted"
  },
  "nodes": [
    {
      "path": "projects/a/envs/dev/apps/backend",
      "kind": "app",
      "key": "backend",
      "name": "Backend",
      "selected": true,
      "selectedBy": "user",
      "action": "update",
      "matchedBy": "id",
      "changes": ["deployment.resources", "settings.envVars"],
      "restart": true,
      "deploy": false,
      "issues": [],
      "notes": []
    }
  ],
  "summary": {
    "create": 0, "update": 1, "unchanged": 0, "keep": 0, "restart": 1, "deploy": 0,
    "blocked": 0, "skipped": 0, "fixable": 0, "warning": 0
  },
  "planHash": "9f2c…"
}
```

`kind` is one of `global`, `project`, `env`, `settings` and `app`. `nodes` holds
every node of the bundle inside the route's scope, selected or not, so the
dashboard can draw the whole tree. Individual settings are not nodes: a
`settings` node's `changes` names the settings that differ, and their issues are
its issues.

### Apply response

The validate response with an `outcome` on each node - `applied`, `failed`,
`skipped` or `unchanged` - the report of what actually happened, and the
deployments queued:

```json
{ "deployments": [ { "appId": "…", "deploymentId": "…" } ] }
```

### The route bounds what can be written

- **Global** writes anything in the bundle. It is the only route that creates
  projects.
- **Project** writes only below `projects/<p>`, where `<p>` is the route's
  project, found in the bundle by id and then by key. A bundle that does not hold
  it is refused with `ERR_SPEC_IMPORT_SCOPE_NOT_IN_BUNDLE`.
- **Env** writes only below `projects/<p>/envs/<e>`, found the same way.

A wider bundle is fine: a global bundle imported on a project's page takes that
project's subtree. A reference that leaves the route's scope - an app imported at
env scope that uses a global certificate - resolves against the target or
becomes an issue. Closure (§5) never reaches outside the route.

### Permissions

- Write access at the route's scope: the access checks export makes for read -
  `GeneralResourceAccessCheck` globally, `ProjectAccessCheck` for a project or an
  env - with `ActionTypeWrite`.
- When the bundle's secrets mode reveals secrets, `AuthorizeSecretReveal` as
  well, the gate export applies to producing such a bundle. Validate compares the
  target's secrets with the bundle's, and "same" or "different" is enough to test
  a guess.
- The two gates template creation applies: granting capabilities needs Write on
  the cluster module, and a mount reaching another app's storage needs Write on
  that app (§6).
- The gate project update applies to changing an existing project's owner (§4).

### Errors

An error means validate could not produce a plan, or apply refused one.
Everything about a single object is an issue in the plan instead.

| error | when |
|---|---|
| `ERR_SPEC_PASSPHRASE_REQUIRED` (exists) | an encrypted bundle, and no passphrase |
| `ERR_SPEC_PASSPHRASE_INVALID` | the passphrase does not open it |
| `ERR_SPEC_BUNDLE_INVALID` | not an archive export produced, or a file in it fails to parse |
| `ERR_SPEC_API_VERSION_UNSUPPORTED` | an `apiVersion` this installation does not read |
| `ERR_SPEC_IMPORT_SCOPE_NOT_IN_BUNDLE` | the route's project or env is not in the bundle |
| `ERR_SPEC_IMPORT_PLAN_CHANGED` | apply: the recomputed `planHash` differs from the one sent |
| `ERR_SPEC_IMPORT_ISSUES_NOT_ACCEPTED` | apply: the plan has issues, and `acceptIssues` is false |
| `ERR_SPEC_IMPORT_BLOCKED` | apply: the plan has a blocked issue |
| `ERR_REQUEST_TOO_BIG` (exists) | the body is over 25 MB |

## 4. Matching

| object | matched by | then by | why |
|---|---|---|---|
| project | `id` | `key` | the key is derived from the name at creation and never changes |
| env | its id, derived as `CalcProjectEnvID(projectID, name)` | `(project, key)` | name and key never change; only colour and position do |
| app | `id` | `(env, key)` | the key never changes |
| collection setting | `id` | `(scope, type, key)` | names change, and the id is what recognises a rename |
| singleton setting | `(scope, type)` | - | there is at most one per scope |

A bundle made on another installation never matches by id - ULIDs do not
collide - and falls through to keys. A target whose database was copied from the
source does match by id, and correctly: they are the same objects.

**The export change.** `ProjectDoc` and `AppDoc` gain `id`, and every entry of a
collection block gains a top-level `id`. Env ids are derived, so `EnvDoc` needs
none. No exported setting type has a top-level `id` in its data today -
`BackupSnapshot` does, and export skips it - and a test keeps it that way.
`ProjectDoc` also gains `owner: { id, email }`. The email is not a secret, so it is
written in every secrets mode, including `omit`.

**The project owner** is a user, and users do not travel, so it is resolved:

1. the user with the owner's `id`;
2. otherwise the user with the owner's `email` - emails are unique among users
   who are not deleted (`idx_uq_users_email`), and `GetByEmail` compares them
   lowercased;
3. otherwise the operator importing, with note `OWNER_NOT_FOUND`.

A candidate counts only if it is an active user, as project update requires of a
new owner; a disabled or pending one falls through to the next step. On the
installation that produced the bundle the id finds the owner even if their email
changed since; elsewhere the email finds the same person under another id.

A new project gets the resolved owner. An existing one is given it only if the
operator may change the owner by hand - an admin, the current owner, or Write on
the Project module, the gate project update applies. Otherwise the owner stays,
and the project takes `OWNER_NOT_PERMITTED`.

**Rules**

- **An id is optional.** A bundle without ids is matched by key, which keeps
  bundles exported before this change importable. Section 7 says what that costs
  them.
- **An id matching an object with a different key** is `KEY_MISMATCH` for a
  project, env or app. Keys never change, so the bundle was edited - usually an
  attempt at cloning - and following the id would rename and overwrite the
  original. The object is skipped. For a setting it is a rename since the
  export, and the setting is updated in place with the bundle's name.
- **A key colliding with a different object that cannot be merged** - a project
  name held by a project with another key - is `NAME_IN_USE`, and the object is
  skipped.
- Unchanged from the export spec: seeded objects are matched by key and merged,
  never duplicated, and `is_default` is compared per `(scope, type, kind)`.

Reservations follow the settings. Domain reservations live in `res_links`, which
the setting repository rewrites whenever it writes a setting, so import gets them
by writing settings through it. Published ports are checked against the services
Swarm runs (`VerifyPortsAvailable`).

## 5. Validate

### An export of the target, diffed against the bundle

Export already builds a bundle in memory before writing it to disk:
`buildBundle` returns a `specmodel.Bundle`, and only `writeBundle` touches a file.
Validate runs `buildBundle` for the target - at the bundle's scope, walking only
what is selected, with the bundle's secrets mode - and diffs the result against
the uploaded bundle. A reference to something outside that walk is resolved by
looking the target object up directly, by id and then by key.

Both sides are then produced by the same code. References are scope paths,
secrets are in the same mode, and labels are filtered and digests stripped the
same way, so the diff compares like with like and needs no comparator per setting
type. Two rules keep it honest:

- `ExternalRef`s are compared by `(type, name, kind)`, never by `id`: ids differ
  between installations for the same object.
- Secret values are compared in memory and never returned.

### Each node

| field | meaning |
|---|---|
| `action` | `create`, `update`, `unchanged`, or `keep` - the object exists and `existing` is `keep` |
| `matchedBy` | `id` or `key`; empty for `create` |
| `changes` | the names of what differs - blocks such as `deployment.resources`, settings such as `sslCerts/localhost@self-signed` - and never their values |
| `restart` | the app has a running service, and a change reaches it: any deployment block other than `source`, or env vars, secrets, config files or routing |
| `deploy` | a deployment will be queued for it under the options |
| `selectedBy` | `user`, or `dependency` when closure pulled it in |
| `issues`, `notes` | §6 |

### Closure

Selecting one app almost always selects too little: its certificate and its
volume live at other scopes. Closure is therefore always on.

- It follows references - `GetRefObjectIDs()` - and never containment. An app
  pulls in the certificate it uses, never its sibling apps.
- It pulls in only what the target lacks. An object the target already has stays
  unselected and the reference resolves to it, so importing one app does not
  update a certificate forty other apps use. The operator can still select it.
- It never reaches outside the route's scope.
- A dependency the operator deselects resolves against the target like an
  external reference, and is `REF_NOT_SELECTED` only if that fails.

### `planHash`

The SHA-256 of:

- the digest of the bundle's bytes;
- the selection and the options;
- every node's `(path, action, restart, deploy)` and its issue codes, sorted.

The target's current values are left out on purpose. A field somebody edits
between validate and apply still ends at the bundle's value, so the operator has
nothing new to decide. What does change the hash - a value newly dropped, a
domain taken in the meantime, a `create` that became an `update` - is exactly
what they would want to see again before going on.

## 6. Issues

### Severity

| severity | effect | apply |
|---|---|---|
| `blocked` | nothing can be written | refused |
| `skipped` | the object is not imported | needs `acceptIssues` |
| `fixable` | part of the object is cleared, and the rest is imported | needs `acceptIssues` |
| `warning` | nothing is cleared, but the result may not be what the operator expects | needs `acceptIssues` |

`warning` is new, and `specmodel.Severity` gains it. A note is not an issue: it
says what import did, and needs no acceptance.

### Codes

| code | severity | raised when | effect | status |
|---|---|---|---|---|
| `SETTING_VERSION_NEWER` | blocked | a setting's `Version` is newer than this installation reads (`ErrDataVerNewerThanSystemVer`) | - | - |
| `KEY_MISMATCH` | skipped | §4 | not imported | - |
| `NAME_IN_USE` | skipped | §4 | not imported | - |
| `TYPE_NOT_IMPORTABLE` | skipped | a setting type with no import policy (§10) | not imported | - |
| `CAPABILITY_NOT_PERMITTED` | skipped | the app gains capabilities, and the operator lacks Write on the cluster module | app not imported | - |
| `SHARED_MOUNT_NOT_PERMITTED` | skipped | a mount reaches the storage of an app outside this import that the operator may not write | app not imported | - |
| `REF_NOT_SELECTED` | fixable | a reference to a bundle object that is not selected and that the target lacks | reference cleared | `pending` |
| `REF_NOT_FOUND` | fixable | a reference to something neither in the bundle nor on the target | reference cleared | `pending` |
| `SECRET_OMITTED` | fixable | §7 | secret left empty | `pending` |
| `OWNER_NOT_PERMITTED` | fixable | an existing project's resolved owner differs, and the operator may not change owners (§4) | owner kept | `active` |
| `NODE_NOT_FOUND` | fixable | a `cluster-volume` pinned to a node the target lacks | pin dropped | `active` |
| `DOMAIN_IN_USE` | fixable | a domain another app has reserved | domain dropped | `active` |
| `PORT_IN_USE` | fixable | a published port another service holds | port dropped | `active` |
| `STORAGE_NOT_EMPTY` | warning | an app being created already has data at its storage path | created on that data | `active` |
| `STORAGE_UNCHECKED` | warning | the storage path of an app being created could not be read | created without knowing | `active` |

The two storage codes come from the check the template preflight uses,
`InspectAppStorage`, which reads volumes on other nodes through their agents. A
path the check could not read is reported as unchecked, never as empty.

Notes:

- `CREDENTIAL_KEPT` - an app's kind credential stays the target's (§7).
- `SECRET_GENERATED` - a created object's omitted secret was generated (§7).
- `OWNER_NOT_FOUND` - no active user has the owner's id or email, so the
  operator importing owns the project (§4).

### Values tied to an installation

Each is decided against the target, value by value:

| value | when the target has it | otherwise |
|---|---|---|
| `Secret.SwarmRef`, `ConfigFile.SwarmRef` | kept | cleared, and applying the app creates the Swarm object again. Not an issue: it is what happens on every other installation, and nothing is lost |
| `cluster-volume.NodeID` | kept | `NODE_NOT_FOUND` |
| image digest | - | export strips it (correction 3) |
| `ssl-cert` material | imported whole | - |

Neither the report nor the dashboard phrases any of this as "same installation"
or "different installation". A node removed at home and a node that never existed
elsewhere are the same finding.

## 7. Secrets

A bundle's `secretsMode` decides what it carries.

**`omit`** - no secret values:

| object | behaviour |
|---|---|
| exists | the target's value is kept. An omitted secret never overwrites a value with nothing |
| created, holding a value HivePaaS owns: `app-kind` database, cache and storage credentials, a project's webhook secret | generated; note `SECRET_GENERATED` |
| created, holding anything else: `secret` settings, registry auth, OAuth, a certificate's private key | left empty; `SECRET_OMITTED`; the object is `pending` |

Generating is right for the values HivePaaS owns because nothing depends on them
yet: an app being created initialises its storage with whatever it is given. The
exception is an app created on storage that already holds data, and
`STORAGE_NOT_EMPTY` reports that.

**`encrypted` and `plaintext`** - the bundle's value is written, except for the
`app-kind` credentials:

| match | behaviour |
|---|---|
| by id: the same app | the bundle's credential, because the data it protects was initialised with it |
| by key: another app with the same key - recreated since the export, or on another installation | the target's credential if it has one, with note `CREDENTIAL_KEPT`, because the data on the target was initialised with the target's |
| created | the bundle's credential |

A bundle without ids matches every app by key, so it always keeps the target's
credential - the safe side of the rule.

Validate reads the target's secrets to compare them, in memory; §3 gives the
permission that requires.

## 8. Apply

Apply runs validate again on its request, and refuses unless:

- the recomputed `planHash` equals the one sent - otherwise
  `ERR_SPEC_IMPORT_PLAN_CHANGED`, and the dashboard shows the new plan;
- nothing is `blocked` - otherwise `ERR_SPEC_IMPORT_BLOCKED`;
- `acceptIssues` is true, if anything is `skipped`, `fixable` or `warning` -
  otherwise `ERR_SPEC_IMPORT_ISSUES_NOT_ACCEPTED`.

Then it runs three phases.

### Phase 1: the database, in one transaction

1. Choose the id of everything to be created, then remap every reference to its
   target id - `RemapRefs`, with its self-check - before the first write. Objects
   of one import can then refer to each other, as the apps of one template
   already do.
2. Global settings.
3. Projects. A new one is created through `projectservice` exactly as the
   dashboard creates one, defaults included, owned by the owner resolved in §4,
   and the bundle's project settings are then matched against those defaults by
   key and update them - the default notification, the webhook. An existing one
   gets its name and note, and its owner under the gate in §4.
4. Envs: their rows, built as project update builds them, then their settings.
5. Apps. New ones go through `appprovisionservice.ProvisionApps`, which creates
   their Swarm services inside the transaction and, with `deployCreated`, their
   first deployments - created, not scheduled. Existing ones get their record and
   settings written; their services are phase 2's.
6. Statuses from §6, and an audit record of type `spec-import` carrying the
   bundle digest, scope, selection, options and counts.

A failure rolls the transaction back, and `Cleanup` removes what `ProvisionApps`
created, as template creation does. Nothing is left behind.

### Phase 2: Swarm, after the commit, for existing apps

For each updated app that has a service:

1. `ServiceInspect`, run the builders (§9) on its spec, and `ServiceUpdate` once:
   one restart, where the settings screens make one update per block.
2. `ApplyAppConfiguration`: environment, config files, secrets, routing,
   scheduled jobs.

A running service cannot be rolled back, which is why this phase comes after the
commit rather than inside it. An app that fails here gets the outcome `failed`
with its error: its configuration is saved, and its service is not updated.
Running the same import again retries it, because the difference between the
bundle and the service is still there.

### Phase 3: deployments

- With `deployCreated`, schedule the first deployments phase 1 created.
- With `deployChangedSource`, create and schedule a deployment for each updated
  app whose `deployment.source` differs.

They go through the task queue like any other deployment.

### Synchronous, on purpose

All three phases run inside the apply request, as creating from a template does.
A background task would have to keep the bundle and the passphrase in
`Task.Args`, which is a database column - exactly what §2 keeps the design from
doing. A very large import can take minutes; the selection lets an operator
import one project at a time.

## 9. Building apps

`specservice.BuildApp` and `appprovisionservice.ProvisionApps` were written for
templates with import in mind, and both carry a `TODO: spec import`. `BuildApp`
builds only what a template may ask for: `CheckBuildable` refuses the rest, and
caps what it accepts at ten secrets or config files, five domains and five
published ports. Import needs everything export writes.

- **The builder registry grows to every block and app-scope setting type export
  writes.** A test holds it there, the way
  `TestBuilderRegistryCoversEveryBuildableBlock` holds the template registry to
  `BuildableBlocks`.
- **Two gates.** `CheckBuildable` stays as it is, for templates. `CheckImportable`
  is new: it checks only that a document is well formed, accepts references once
  they are remapped, and has none of the template caps.
- **A builder replaces its block; it never appends.** That is what lets the same
  builder run on an existing service's spec and converge on the bundle.
- **Absence is meaningful inside `deployment`.** A sub-block missing from a
  present `deployment` means empty, and clears the service's. A missing
  `deployment` - an app never deployed at the source - leaves the service alone.
- **`container.image` is never written to a service.** It records what was
  running at export; the image changes only through a deployment.
- **Import is a mode of `BuildApp`.** `BuildAppReq.Import` swaps `CheckBuildable`
  for `CheckImportable` and `PresentBlocks` for `ImportBlocks`;
  `deployment.container` and `deployment.service` are built only in it. What the
  settings screens keep for HivePaaS is kept here too: the labels
  `ApplyUserLabels` protects, the placement constraints HivePaaS derived, the
  network an app was created on, and HivePaaS's log driver when none is named.

**Storage travels in the form the builder reads.** A mount into a volume's
directory for an app of its environment is exported the way a template and the
storage screen write it: the volume by its scope path - or, when the export does
not hold it, by `external` - the path below the app's directory, and `sourceApp`
when the directory is another app's. Every other mount travels in
`storage.dockerMounts` exactly as Docker holds it, and the shared-memory mount in
`resources.memory.shmSize` alone. Two things a managed mount can carry do not
travel: a per-mount driver override and labels, because export cannot tell them
from the volume's own, which a mount inherits when it names none. Building the
volume's mount again inherits them again.

Why not move the `apply*` functions of `appsettingsuc` into a service, as the
export spec proposed:

- the builder is already a service, shared with templates;
- one `ServiceUpdate` per app instead of one per block, so one restart;
- there is one mapping from bundle to Swarm to test, against the one export uses
  from Swarm to bundle: build a spec from a document, map it back with
  `swarm_map`, and require the same document.

## 10. Projects, envs and scope settings

**Project creation moves to `projectservice`.** The defaults a new project gets -
its env rows, tags, webhook, default notification and default volume - are built
by `preparePersistingProject*` in `projectuc`, and one usecase may not call
another. In the import plan they move to `projectservice`, where
`projectuc.CreateProject` calls them, unchanged in behaviour, and import calls
them too. So does the env-row construction of project update.

**Settings are written through import policies**, a registry keyed by setting
type, beside the builders in `specserviceimpl`:

- the default policy parses with the type's registered parser, validates, and
  writes the row;
- a type with work beyond its row registers a policy that does that work;
- every type with an export policy has an import policy, or an explicit skip with
  a reason, and a test holds that. A type in a bundle with neither is
  `TYPE_NOT_IMPORTABLE`.

This is the export registry's safety rule turned around: no setting type is
imported because somebody forgot it could not be.

**The template checks are shared once import needs them.** Template creation
checks a request in `apptemplateuc` - five checks and the storage preflight - and
each reads what a document asks for, then asks a service. The reading moves to
`specmodel` in the import plan, where import becomes its second caller: the
ports a document publishes, the domains it answers at, the capabilities it
grants. The service calls stay with each caller, because the two use them
differently: template creation refuses on the first finding, while import reports
every one and must leave out the app it is updating. Storage and shared mounts
wait for the builder-coverage plan, which decides how export represents a mount
(§9). The permission halves stay in each usecase, which is where permissions are
decided. The fifth check, `checkAppRefs`, resolves template parameters, which
import does not have; its references are resolved by §4 and §5.

## 11. Dashboard

**Where.** Beside each export screen: Operations → Import globally, and a
project's Operations → Import, whose scope follows the environment picker the way
project export does - with no environment selected, the project; with one
selected, that env.

**Steps.**

1. Pick the file, and ask for the passphrase on `ERR_SPEC_PASSPHRASE_REQUIRED`.
2. The plan:
   - what the bundle is: scope, export date, version, secrets mode;
   - the tree - global, projects, their settings, envs, their settings, apps -
     with a checkbox, an action badge and a restart badge on each node, a mark on
     what closure pulled in, and issues inline;
   - the three options;
   - the summary.
3. Changing the selection or an option validates again, debounced.
4. Confirm. The button says what it accepts: with no issues it simply imports;
   with some it names how many things will be left out or are warned about; with
   a blocked issue it is disabled.
5. The result: each node's outcome, failed apps with a retry that re-runs the
   same import, and links to the deployments queued.

On `ERR_SPEC_IMPORT_PLAN_CHANGED` the new plan replaces the old one, and the
operator confirms again.

The file stays in the page's memory for the whole flow. Reloading the page means
picking it again, which is the price of the server keeping nothing.

## 12. What changes

| where | change |
|---|---|
| `specmodel` | ids on `ProjectDoc`, `AppDoc` and collection entries; `owner` on `ProjectDoc`; `SeverityWarning`; the issue codes; the selection, options, plan and node types |
| `specserviceimpl` | writing ids and the project owner; the target export for validate; owner resolution; matching, diff, closure, issues and `planHash`; the three apply phases; the extended builder registry and `CheckImportable`; import policies; the document halves of the template checks |
| `specservice` | `ValidateImport` and `ApplyImport` |
| `specuc` | the import usecases, with their permissions and audit |
| `spechandler`, routers | six routes, and the body limit |
| `projectservice`, `projectuc` | project defaults and env rows moved down |
| `apptemplateuc` | reads documents through `specmodel`; keeps its own service calls |
| `hperrors/errors_spec.go` | the new errors |
| `base/audit.go` | `AuditLogTypeSpecImport` |
| `hivepaas-dashboard` | the import screens, global and per project |

## 13. Testing

- **Builder round trip.** For a fixture carrying every block, build a spec from
  the document, map it back with `swarm_map`, and require the same document.
- **Registry coverage.** Every block and app-scope setting type export writes has
  a builder; every setting type with an export policy has an import policy or an
  explicit skip.
- **Collection ids.** No exported setting type has a top-level `id` in its data.
- **Export, import, export.** Export a scope, import it into an empty database,
  and export again: the two bundles are equal but for ids and the values §6
  resolves against the target. This is the migration path, end to end.
- **Idempotence.** Import a bundle into the installation that produced it: every
  node is `unchanged`, nothing restarts, nothing deploys, and a second validate
  returns the same `planHash`.
- **Matching.** A setting renamed since the export is updated, not duplicated. A
  bundle without ids is matched by key. A project whose key was edited in the
  bundle is `KEY_MISMATCH`, and the original is untouched.
- **Project owner.** Found by id even after their email changed; found by email,
  in another case, when the id is unknown; neither gives the operator importing,
  with `OWNER_NOT_FOUND`. A disabled user found by id falls through. An existing
  project whose owner would change, imported by somebody who may not change
  owners, keeps its owner with `OWNER_NOT_PERMITTED`.
- **App-kind credentials survive a round trip.** Export an encrypted bundle and
  import it again: the credential is byte-identical. Delete the app, create it
  again with the same key, and import: `CREDENTIAL_KEPT`, and the new app's
  credential is untouched. Import an `omit` bundle into an empty database: the
  credential is generated, and noted.
- **Omitted secrets never erase.** An `omit` bundle imported over an existing app
  leaves every secret as it was.
- **Selection.** Each of the six cases against one fixture, asserting exactly the
  intended nodes are written - in particular that `projects/a/settings` writes
  the project and its settings and no env.
- **Closure.** One app pulls in the certificate it uses when the target lacks it,
  and leaves it unselected when the target has it. It never pulls in a sibling
  app, and never reaches outside the route's scope.
- **`REF_NOT_SELECTED` against `REF_NOT_FOUND`**, as the export spec describes,
  with the dependency deselected.
- **Issues and status.** `NODE_NOT_FOUND` unpins the volume and leaves it
  `active`. `DOMAIN_IN_USE` drops that one domain and leaves routing `active`.
  `REF_NOT_FOUND` leaves its setting `pending`.
- **`planHash`.** It is stable across two validates of the same request. A domain
  taken between validate and apply makes apply return
  `ERR_SPEC_IMPORT_PLAN_CHANGED`; editing an unrelated field of an app in between
  does not.
- **Acceptance.** A plan with issues is refused while `acceptIssues` is false; a
  plan with a blocked issue is refused whatever the flag says.
- **Phase 1 rollback.** A failure after `ProvisionApps` leaves no row and no Swarm
  service behind.
- **Phase 2 partial failure.** When one app's `ServiceUpdate` fails, the others
  are applied, the report names it, and importing again completes it.
- **`existing: keep`.** Nothing that exists changes; only missing objects are
  created.
- **Permissions.** An encrypted bundle without `AuthorizeSecretReveal` is refused
  at validate. Capabilities without Write on the cluster module are
  `CAPABILITY_NOT_PERMITTED`.
- **Route scope.** A global bundle imported at a project's route writes only that
  project. A bundle without the route's project is
  `ERR_SPEC_IMPORT_SCOPE_NOT_IN_BUNDLE`.
- **Body limit.** A body over 25 MB is `ERR_REQUEST_TOO_BIG`.

## 14. Order of work

Five plans, each shippable without the next:

1. **Export ids and owner**
   (`docs/superpowers/plans/2026-09-23-spec-export-ids-and-owner.md`). Export
   writes the ids and the project owner §4 needs. Nothing reads them yet, but
   every bundle exported from then on can be matched exactly.
2. **Export references and storage**
   (`docs/superpowers/plans/2026-09-24-spec-export-references-and-storage.md`).
   A reference that leaves the export becomes an external one (correction 7),
   and storage travels in the builder's form (§9).
3. **Builder coverage.** Deployment blocks:
   `docs/superpowers/plans/2026-09-24-spec-import-builders.md` - import mode,
   `CheckImportable`, builders that replace rather than append, and a round trip
   through export's mapping. In import mode every setting, the source included,
   is built from the row export wrote beside its data:
   `docs/superpowers/plans/2026-09-24-spec-setting-rows.md`.
4. **Import**: `ValidateImport` and `ApplyImport`, the usecases, handlers and
   routes, the errors and the audit type. Project creation moves to
   `projectservice`, and the template checks' reading of a document to
   `specmodel`, in this plan - where import becomes their second caller, so both
   callers shape what is shared. Validate, first part:
   `docs/superpowers/plans/2026-09-24-spec-import-validate.md` - reading a
   bundle, matching, the diff, restart and deploy, the structural issues, the
   plan hash, and the validate endpoint at the three scopes. Validate, second
   part: `docs/superpowers/plans/2026-09-24-spec-import-references.md` -
   references and closure, `keep` creating the settings a scope lacks, and the
   records a selected object needs. Availability issues, owner and credential
   notes, and apply follow in their own plans.
5. **Dashboard**: the import screens.

---

## Later, not now

- Cloning under a new key.
- Offering to empty storage that already holds data, as the template preflight
  does.
- A diff per field in the plan, rather than the names of what changed.
- Background import, which first needs somewhere other than the database to keep
  a bundle.

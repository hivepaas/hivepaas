# Setting Mounts, the One Way to a File

Setting mounts are a second way to put a file in an app's container. Secrets
and config files had the first: a `swarmRef.file` on each. Two ways to say the
same thing is one too many. The old one could not mount a project's secret,
which confused people who could see that secret from the app.

This design makes setting mounts the only way. Secrets and config files become
sources of setting mounts, and lose `swarmRef`. It amends
`2026-09-25-setting-mounts-design.md`, whose §§ are cited here as "base §".

HivePaaS is in development: nothing here is carried over from installations
that used `swarmRef`.

---

## Decisions

1. **Secrets and config files are sources.** `secret` offers `value`,
   `config-file` offers `content`. A base64 setting is decoded before its file
   is written.
2. **"Sensitive" is two things.** A part may be *stored as a secret* (a Docker
   secret rather than a config) and may be *gated* (mounting it asks for Reveal
   Secrets). A secret's value is stored as a secret and is not gated: the app
   already reads it through `${secrets.NAME}`. An SSL key, an SSH key, a
   password and an htpasswd are both.
3. **Sources are found as settings are.** The app's own, its parent app's, its
   env's and its project's when inheritable or shared into the app, as base §4
   already reads them. The environment's reading of secrets, which ignores
   `inheritable`, is made to agree in a separate task.
4. **No `swarmRef` on secrets and config files.** Not in the entity, the API,
   export or the dashboard. Docker objects are no longer made for them.
5. **Templates keep their syntax.** In a template - and only there - a secret's
   or config file's `swarmRef.file` is shorthand: building the app turns it
   into a setting mount entry. The 45 templates that mount files are not
   rewritten.
6. **Entries can be inheritable.** `Setting.Inheritable`, false by default,
   chosen by whoever creates the entry. Previews and clones take inheritable
   entries only.

## 1. Parts

Base §2's table gains two rows and a column:

| source type | part | required | stored as secret | gated |
|---|---|---|---|---|
| `secret` | `value` | yes | yes | |
| `config-file` | `content` | yes | | |
| `ssl-cert` | `certificate` | yes | | |
| | `privateKey` | yes | yes | yes |
| | `caCertificate` | | | |
| `ssh-key` | `privateKey` | yes | yes | yes |
| | `publicKey` | | | |
| `basic-auth` | `username` | yes | | |
| | `password` | yes | yes | yes |
| | `htpasswd` | yes | yes | yes |

- *Stored as secret* decides Docker secret or Docker config (base §2's
  "sensitive" for storage).
- *Gated* is base §7's "sensitive": mounting it passes `AuthorizeSecretReveal`
  when the set of (source, gated part) pairs an app mounts grows.
- The dashboard marks both, and locks only gated parts for someone who may not
  reveal.

## 2. What goes

- **The fields.** `entity.Secret.SwarmRef`, `entity.ConfigFile.SwarmRef`,
  `SwarmSecretRef` and `SwarmConfigRef`.
- **The API.** The `swarmRef` of the secret and config file DTOs, requests and
  responses.
- **The Docker side.** `clustersecretservice`'s making, updating and removing
  of Docker secrets and configs for apps, including attaching a parent's to its
  previews, and the `makeRoom` of plan 2. The service goes if nothing else
  uses it.
- **Callers.** Provisioning's `applySwarmFiles`, import's `updateSwarmFiles`,
  the secret and config file use cases' Docker calls, and deletion's removal of
  referenced secrets and configs.
  - An app's secrets and configs in Docker are now only setting mounts', which
    `settingmountservice.RemoveApp` removes.
- **Path checks.** Paths are checked among entries only: base §1's "unique
  among the app's entries, secrets and config files" becomes "among the app's
  entries". `CheckMountPaths` on secrets and config files goes.
- **The dashboard.** The mount fields of the Secret and Config File forms, and
  their Mountpoint column.
- **Rows already stored.** A `swarmRef` in a row, or in a bundle being
  imported, is ignored when read.

## 3. Templates and HivePaaS's own apps

- **The shorthand.** In the buildable subset (`specmodel/buildable.go`), a
  secret or config file may still carry `swarmRef: {file: {name, uid, gid,
  mode}}`. Building the app creates the setting without it, and an
  `app-setting-mount` entry:
  - the key is the setting's key, lowercased and made a valid entry key;
  - the source is the new setting;
  - one file: part `value` or `content`, path `file.name`, and uid, gid and
    mode as given;
  - `Inheritable` is the setting's own.
- **Inheritable, in templates.** Secrets and config files may carry
  `inheritable: true`, which sets both the setting and its entry. Previews read
  a source through their own scope, so an inheritable entry needs an
  inheritable source.
- **The 45 templates.**
  - Each config file that mounts a file gets `inheritable: true`, so that a
    preview still has it.
  - Secrets stay false; the template's author turns them on where a preview
    needs them.
- **Registry and logging.** Both build their apps from documents (the
  registry's `app.yaml.tmpl`, logging's `appdoc.go`), so the shorthand serves
  them. `systemappservice/secrets.go`, which sets `SwarmRef` on entities, makes
  entries instead.
- **Export.** Export writes entries (base §9), not the shorthand. The shorthand
  exists only for templates.

## 4. Previews and clones

**Previews:**
- A preview app resolves the inheritable entries of its parent, read live, as
  if they were its own.
- It has none of its own: the entry screen is the parent's.
- Its Docker objects carry the preview's name and label, so deleting it removes
  only its own.
- The refresh task (base §5) refreshes an app's previews with it, both when a
  source changes and when an entry does.
- **What this hands out.** A preview runs the code of a pull request. An
  inheritable entry's files - a private key among them - reach whoever can open
  one. Marking an entry inheritable is consenting to that.

**Clones:**
- A clone copies the source app's inheritable entries.
- **Sources.** An entry whose source is a secret or config file of the source
  app is pointed at the clone's copy of it, when `CloneSecrets` or
  `CloneConfigFiles` copied it. Otherwise the entry is left out: the clone
  cannot see the source app's settings. A source outside the app keeps its id.
- **Gated parts.** When the request is made, while its session is there,
  inheritable entries with gated parts pass `AuthorizeSecretReveal` once,
  recorded.
  - Denied, the clone goes ahead without those entries.
  - The response says which were left out.
- **The rest.** Entries that are not inheritable are not copied. Base §6's
  clone and preview rows are replaced by this section.

## 5. Plans

1. **Backend, sources.**
   - The `secret` and `config-file` sources, and the stored/gated split.
   - `swarmRef` removed from secrets and config files, and what goes with it
     (§2).
   - The template shorthand and `inheritable` in templates (§3), and
     `systemappservice`.
2. **Backend, inheritance.**
   - `inheritable` on entries through the API.
   - Previews resolving their parent's, and the refresh reaching them.
   - Clones copying, with the gate at request.
3. **Templates.** `inheritable: true` on the config files that mount files, in
   `app-templates`.
4. **Dashboard.**
   - The Setting Mounts screen of `plans/2026-09-25-setting-mounts-dashboard.md`,
     with the two source types, the `inheritable` switch and gated parts locked.
   - The mount fields removed from Secrets and Config Files.
   - The clone result says which entries were left out.

## Not in this design

- **The environment's reading of secrets** agreeing with settings' visibility
  (Decision 3). A task of its own.
- **Moving data from `swarmRef` to entries.** Nothing is carried over
  (development).

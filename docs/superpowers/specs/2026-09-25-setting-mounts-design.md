# Setting Mounts

An app's secrets and config files are mounted into its container, and a change
to one reaches the container on its own. Their values are their own, though:
nothing lets a file in the container follow another setting. An app that
terminates TLS itself needs the certificate HivePaaS already manages, and today
someone has to copy it into a secret by hand and do it again at every renewal.

This design lets an app mount parts of another setting - a certificate, a key,
a basic auth pair rendered as htpasswd - as files that follow that setting from
then on. TLS passthrough is its first user: a domain passed through to the app
brings its certificate with it.

---

## Decisions

1. **A new app setting, not a linked secret.** Secrets and config files keep
   their meaning. A linked value would reach every place a secret is read -
   environment variables, reveal, export modes, encryption at rest, inherited
   scopes - and each would have to learn that the value lives elsewhere.
2. **One entry, one source, many files.** Every file of an entry comes from one
   setting, and all of them are replaced in one service update. A certificate
   and its key are never seen apart.
3. **Parts, not fields.** A source type offers named parts, each computed from
   its data. A field is the simplest part; a virtual part such as `htpasswd` is
   computed. The engine does not tell them apart.
4. **Rotated by what goes in, not by what comes out.** A part's file is replaced
   when its inputs or its renderer change. bcrypt salts every hash afresh, and
   comparing output would restart the app on every refresh.
5. **Handing out a sensitive part is revealing it.** Whoever controls the
   container can read the file, so mounting a private key or a password takes
   what revealing it takes.
6. **TLS passthrough is derived, never stored as an entry.** The routing
   settings are the one description of it; an entry HivePaaS made would be a
   second one that a person could edit or delete.
7. **A setting that cannot be used is not in the container.** A disabled entry,
   a disabled source, or an empty required part leaves no file behind.

## 1. The setting

A new collection type, `app-setting-mount`, in app scope only and never
inherited. In a document it is `settings.settingMounts`, keyed like secrets:

```yaml
settings:
  settingMounts:
    tls:
      source: <reference to the source setting>
      files:
        - {part: certificate,   path: /etc/app/tls/cert.pem}
        - {part: privateKey,    path: /etc/app/tls/key.pem, mode: "0400"}
        - {part: caCertificate, path: /etc/app/tls/ca.pem}
```

| field | meaning |
|---|---|
| key | The entry's name: `a-z0-9-`, and never `tls`, which TLS passthrough uses (§3). |
| `source` | A setting reference, as every setting reference is: written to `res_link`, remapped by import, and a source still linked cannot be deleted (`ERR_SETTING_IN_USE`). |
| `files[].part` | A part the source's type offers (§2). Each part at most once per entry. |
| `files[].path` | An absolute, clean path. Unique among the app's entries, secrets and config files, and outside `/run/secrets/tls`. |
| `files[].uid`, `gid`, `mode` | Optional. The defaults are those of secrets and config files. |

An entry has at least one file. Its source must be of a type §2 lists and
visible from the app's scope.

## 2. The parts registry

A table in code, one row per part. Adding a source type or a part is adding
rows; the engine does not change.

| source type | part | required | sensitive | notes |
|---|---|---|---|---|
| `ssl-cert` | `certificate` | yes | | as stored: the leaf and its chain |
| | `privateKey` | yes | yes | |
| | `caCertificate` | | | no file when empty |
| `ssh-key` | `privateKey` | yes | yes | |
| | `publicKey` | | | no file when empty |
| `basic-auth` | `username` | yes | | |
| | `password` | yes | yes | |
| | `htpasswd` | yes | yes | virtual: `username:<bcrypt>`, from `pkg/htpasswd` |

Each part declares:
- **render**, which turns the source's data into the file's bytes;
- **inputs**, the fields render reads;
- **version**, raised whenever render's output changes shape;
- **required**, **sensitive**.

A sensitive part becomes a Docker secret, any other a Docker config. A part
derived from a sensitive field is sensitive.

The **rotation key** of a file is an HMAC of its inputs' values and its part's
version, keyed with the app secret. It is what decides whether the file is new
(§4). It is keyed because it ends up in a name anyone who can list secrets
reads, and an unkeyed hash of a password is a way to guess it. Replacing the app
secret renames every mounted object once, at its next refresh.

## 3. Names and labels

Names follow the convention secrets and config files already use,
`GlobalKey + "_" + name`, lowercased:

```
<GlobalKey>_mount_<entry>_<part>_<hash8>      an entry's file
<GlobalKey>_tls_<part>_<hash8>                TLS passthrough (§8)
```

`hash8` is the rotation key. Docker caps names at 64 characters and a
`GlobalKey` can be longer; a name that would be is shortened by cutting the
`GlobalKey` and appending eight characters of its hash, which keeps it unique.
Ordinary secrets and config files have the same limit and are not changed
here.

Labels follow the `hivepaas.<object>.<field>` convention, camelCase:
- `hivepaas.app.id` - the app, as elsewhere;
- `hivepaas.settingMount.entry` - the entry's key, or `tls`;
- `hivepaas.settingMount.part` - the part.

The labels the Docker API design introduced are renamed to the same
convention in this work: `hivepaas.docker-api.app`, `.socket` and `.network`
become `hivepaas.dockerApi.app`, `.socket` and `.network`. Nothing with the old
names has been released.

## 4. The engine

A new service, `settingmountservice`:

- **`Resolve(app)`** is the files the app should have: those of its active
  entries whose source is active and whose required parts all render, and
  those of TLS passthrough (§8). Each comes with its path, its bytes, whether it
  is sensitive, and its rotation key.
- **`ApplyToService(app, spec)`** brings a service spec to that set:
  1. for each file, the secret or config of its name is used when Docker has
     it, and created when not;
  2. the spec's references labeled `hivepaas.settingMount.entry` are replaced
     by those of the set, in the one update the caller makes;
  3. once the update is made, the secrets and configs of the app that no
     reference names any more are removed, retried while a task still holds
     them.

It is called:
- when an entry is saved;
- when routing settings are applied (§8);
- on every deployment, as `dockerAPIService.ApplyToService` is, which puts
  right a refresh that failed;
- by the refresh task (§5).

## 5. When a source changes

Every path that writes a source setting records a `task:setting-mount-refresh`
naming it, in the same transaction, and the task runs once it commits:

| path | how it writes today |
|---|---|
| the settings screens, content and status | the settings use case, through `settingeventservice.OnUpdate` / `OnUpdateStatus` |
| SSL renewal | `settingRepo.UpsertMulti`, after the renewal |
| SSL obtain (`task:ssl-obtain`) | `settingRepo.Update` |
| import | phase 1 records it; it is scheduled with the other tasks after the commit |

The task finds the apps that read the source through `res_link` - an entry
pointing at it, or routing settings whose TLS passthrough uses it - and calls
`ApplyToService` for each. An app that fails is written to the task's log and
the others go on.

## 6. Lifecycle

| event | result |
|---|---|
| entry disabled, source disabled, or a required part empty | the entry's files are removed; they come back when it can be used again, and the screen says why they are gone |
| entry deleted | its files are removed |
| source deleted | refused while linked, as for any reference |
| app deleted | every secret and config labeled `hivepaas.app.id=<app>` and `hivepaas.settingMount.entry` is removed |
| app cloned, preview app created | references copied from the source app are dropped, and the new app is resolved on its own |

## 7. Permissions

Mounting a sensitive part passes `AuthorizeSecretReveal`: the operator's
Return Secrets Via API switch, and the Can Reveal Secrets capability (which an
administrator has). The attempt is recorded, allowed or denied.

It is asked only when the set of (source, sensitive part) pairs an app mounts
grows:
- an entry gains a sensitive part, or an entry with one changes its source -
  asked;
- a sensitive part removed, an entry disabled or deleted - Write on the app.

An entry of plain parts only takes Write on the app, and a source visible from
its scope.

## 8. TLS passthrough

**What is mounted.** The first active domain with `tlsPassthrough` and an SSL
certificate gives the app:
- `/run/secrets/tls/cert.pem` - `certificate`;
- `/run/secrets/tls/key.pem` - `privateKey`;
- `/run/secrets/tls/ca.pem` - `caCertificate`, when it has one.

**Consent.** The routing settings gain `tlsMountedCert`, the certificate whose
mounting passed §7's gate. Files are mounted only while the certificate the
domains ask for is that one.
- Saving routing settings that ask for another certificate passes the gate,
  and on success records it.
- Apps configured before this work have none, so nothing is mounted until
  someone passes the gate and saves. The screen says so.
- A renewal keeps the certificate's id, so its new content is mounted without
  asking again: renewing is not granting.
- Turning passthrough off removes the files and clears `tlsMountedCert`.
  Turning it on again asks again.
- Reading the routing settings says whether the certificate is mounted, and
  when not, whether the reader may mount it - what the screens of §10 show.

**Kind settings.** The database and cache kind screens set passthrough and the
certificate on the first domain (`kind_settings_update.go`), and pass the same
gate. Templates cannot set either.

## 9. Export and import

- **Export** writes entries as they are - a source reference and files - and no
  value. `tlsMountedCert` is not exported: consent belongs to the installation
  that gave it.
- **Import** asks §7's gate for what it would mount. Validate checks without
  recording; apply records.
  - An entry with a sensitive part the caller may not reveal is skipped, with
    `SETTING_MOUNT_NOT_PERMITTED` (severity skipped).
  - Routing with passthrough and a certificate is imported either way. Without
    the gate the certificate is not mounted, and a warning says so.

## 10. Dashboard

- **App settings → Config & Data → Setting Mounts.**
  - A list of entries, each with its state and, when nothing is mounted, why.
  - A dialog to create or edit one: the source type, then the setting, then a
    row per file with the path suggested. Sensitive parts are marked, and
    locked with the reason for someone who may not reveal.
- **Routing and kind settings.** Under a passthrough domain with a certificate:
  "mounted at /run/secrets/tls", or "not mounted - saving needs the Reveal
  Secrets permission".

## 11. Testing

- **Registry.** Each part renders from fixture data; `htpasswd` verifies with
  bcrypt; a rotation key changes with an input or a version and not otherwise,
  `htpasswd` included.
- **Engine.** Against a fake Docker: reuse by name, creation, one update that
  swaps references, removal of what nothing references, the lifecycle rows of
  §6.
- **Triggers.** Each path of §5 records a refresh task.
- **Permissions.** The widening rule of §7, on the screen, in kind settings, in
  routing and in import.
- **Real Docker, skipped without it.** A secret mounted at a path outside
  `/run/secrets`; a rotation that replaces a certificate and key in one update.

## 12. To verify before building on it

- **Secrets outside `/run/secrets`.** Verified on Docker 29.8: a secret takes an
  absolute target, and its mode is kept. The engine's real-daemon test mounts a
  key at `/etc/app/tls/key.pem` and renews it.
- **bcrypt in htpasswd.** Traefik, Apache, Caddy and HivePaaS's registry read
  it; nginx does where the system's crypt is libxcrypt, the default on current
  distributions. The part's description says so.

## 13. Plans

1. **Backend, the engine.** The setting type, the registry, names and labels,
   `Resolve` and `ApplyToService`, the refresh task and its triggers, the
   lifecycle, and the Docker API label rename.
2. **Backend, permissions and surfaces.** §7's gate on the entry endpoints,
   export and import, clone and preview apps.
3. **Backend, TLS passthrough.** `tlsMountedCert`, routing and kind settings,
   and the gate on both.
4. **Dashboard.** The Setting Mounts screen, and the state under passthrough
   domains.

## Not in this design

- **Other source types** - registry credentials, cloud storage keys and the
  rest. Each is rows in §2's table when it is wanted.
- **Names over 64 characters for ordinary secrets and config files.** They have
  the limit today; §3 only keeps the new names within it.
- **Linking from a secret or config file.** Decision 1.

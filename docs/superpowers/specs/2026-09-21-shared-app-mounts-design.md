# Sharing one app's storage with another

**Status:** draft, awaiting review

## 1. The problem

A file manager, a backup tool, an importer: each of them is an app whose whole purpose is to
work on another app's files. Today none of them can. The volume model is not what stops
them - a volume is a project-scoped setting, `Inheritable: true`, and every app in the
project may mount it. What stops them is one function,
[calcMountSubpath](../../../hivepaas_app/service/volumeservice/volumeserviceimpl/app_mounts.go):

```go
case base.ObjectScopeProject:
    subpath = fmt.Sprintf("%v/%v", app.ProjectEnv.Key, app.Key)
...
subpath = filepath.Join(subpath, mnt.VolumeOptions.Subpath)   // the request is appended
```

The prefix comes from the app doing the mounting, and whatever subpath the request asks for
is joined **under** it. So a mount can never name anything but the mounting app's own
directory. That is a good rule - every app is confined to its own data by construction - but
it is absolute, and there is no way to say the one thing a file manager needs to say.

There is no workaround either: bind mounts are not in the buildable subset, and a second
volume does not help, because the prefix follows the app rather than the volume.

## 2. Decisions

| Question | Decision | Why |
|---|---|---|
| How wide is this? | One app names one other app's directory | "Mount the data of app X" can be checked, logged and shown. A general "mount the volume root" grants whatever happens to be in that volume later, which nobody can review |
| What actually changes? | Whose prefix `calcMountSubpath` uses | Everything downstream - volume lookup, bind rewrite, pins, permissions - stays as it is |
| Where is the reference stored? | Nowhere new | The swarm service spec is already the source of truth for mounts; `GetAppStorageSettings` inspects the service and transforms it back. Adding a parallel record would create two truths that drift |
| Then how is a foreign mount recognised later? | Derived from the resolved subpath | A mount is the app's own if its subpath starts with the app's own prefix. Deletion must not depend on a flag somebody forgot to set, or that a hand-edited service lacks |
| Read or write? | Read unless the request says otherwise | The reference is a nested object; `write` absent means read-only. A boolean with a default would make "not stated" and "stated false" the same thing |
| Who may do it? | Write on the target app | The object is concrete, so the check can be too. Write on the mounting app is already required to reach the screen |
| What if the target app is deleted? | The mount is reported as dangling; the app keeps running | Refusing to delete an app because something watches its files makes people stuck. The directory stays where it is, and the borrower keeps seeing it |
| Same environment only? | Yes | Cross-env sharing is a different conversation about what an environment is for. Nothing here forecloses it |
| Data copied on clone? | Never | The data is not the cloner's. The reference is carried over when the cloner may have it, and dropped with a warning otherwise |

## 3. The request

`appsettingsdto.Mount` gains one optional field, used by both get and update:

```go
type Mount struct {
    // ... unchanged
    SourceApp *MountSourceApp `json:"sourceApp,omitempty"`
}

// MountSourceApp says the directory this mount reaches belongs to another app.
type MountSourceApp struct {
    AppID string `json:"appId"`
    // Write is absent for the ordinary case: seeing another app's files is one
    // decision, changing them is another.
    Write bool `json:"write,omitempty"`
    // Name and Dangling are filled in on read only.
    Name     string `json:"name,omitempty"`
    Dangling bool   `json:"dangling,omitempty"`
}
```

`Source` still names the volume, because the target app may mount several and nothing can
guess which one is meant. `VolumeOptions.Subpath` still applies, now relative to the target
app's directory rather than the caller's.

## 4. Writing one

`UpdateAppStorageSettings` passes `SourceApp` through to `volumeservice.AppMountReq`.
`BuildAppMounts` resolves it before building the mount:

1. Load the named app. It must be active and in the same project env as the mounting app.
   Naming the app itself is accepted and means the same as leaving the field out.
2. Check `permission.AppAccessCheck{ProjectID, ProjectEnv, AppID: <target>, Action: Write}`.
   Refuse with `ErrUnauthorized` and a message naming the app.
3. Compute the prefix with the target app in place of the caller:
   `calcMountSubpath(targetApp, mnt, volumeSetting)`.
4. Set `ReadOnly = !SourceApp.Write`, ignoring the top-level `readOnly` for this mount, so
   there is one place the answer comes from.

The volume itself is looked up exactly as before, in the *mounting* app's scope. A volume
the caller may not mount stays refused, whoever owns the directory inside it.

The check runs when the mount is created or changed. `BuildAppMounts` rebuilds only the
mounts a request touches - unchanged ones arrive as `Kept` - so saving an unrelated change
does not re-check a mount already in place, and **revoking somebody's access to the target
app does not unmount anything**. Revocation is §13.

`calcMountSubpath` changes signature from `(app, mnt, setting)` to
`(prefixApp, mnt, setting)`; the only caller that passes something other than the mounting
app is this one.

## 5. Reading it back

`GetAppStorageSettings` inspects the service and resolves each mount back to volume and
subpath - for a managed local volume that means undoing the bind rewrite, which
`bindStorageTarget` already does. Add to it:

- if the subpath starts with the app's own prefix, the mount is the app's own and
  `sourceApp` is absent, exactly as today;
- otherwise, match the prefix against the apps of the environment by key and report
  `sourceApp: {appId, name, write: !readOnly}`;
- if no app matches, report `sourceApp: {name: "<the prefix>", dangling: true}` - the owner
  was deleted, and the UI says so rather than showing a path with no explanation.

This needs the environment's apps in `StorageSettingsTransformInput`; the usecase already
loads `ProjectEnv` and can list them in the same query it makes for the app.

## 6. Deleting an app

This is where the feature can destroy data, and the rule is deliberately not "skip the
mounts that carry the new field":

> **An app deletes a directory only when the resolved subpath starts with its own prefix.**

`RemoveAppStorage` already computes the subpath per mount and already refuses a bind that
points at a volume's root. The ownership test goes in the same place, before
`removeStorageTarget`. Because it is computed from the path rather than read from a field,
it holds for mounts written before this feature existed, for mounts edited outside HivePaaS,
and for a service spec restored from a backup.

The prefix to compare against is `calcMountSubpath(app, &AppMountReq{}, volumeSetting)` - the
same function, with no request subpath - so the two can never disagree about what the app
owns.

## 7. Cloning an app

`clone_4_volumes` rewrites each mount's directory from the source app's name to the copy's.
A mount reaching another app's directory has no name of its own to rewrite -
`calcVolumeMountSubpath` finds nothing to replace and `calcBindMountPath` finds no
`<project>/<env>/<app>/` to cut - so both already answer false and the mount is skipped.
Foreign mounts are therefore dropped from a clone, and their data is never copied, with no
new code. That is also the right answer: a copy that silently kept access nobody checked
would be a way to launder permission. Carrying the reference over for a cloner who does hold
Write on the target app is §13.

A test pins this, because it is now load-bearing rather than incidental.

## 8. What this does not cover

- **A volume scoped to a single app.** The volume lookup runs in the mounting app's scope, so
  a volume belonging to app B is invisible to app A and the mount is refused. The default
  project volume - where app data actually lives - is project-scoped, so the case that
  matters works. Widening scope resolution is a separate decision.
- **Across environments or projects.** Refused, by the same-env check.
- **A pinned volume on two nodes.** Both apps end up pinned to the volume's node;
  `refuseConflictingVolumePins` already catches a contradiction. Its message should name the
  other app, since sharing makes this collision likelier than it was.
- **User and group ids.** `ensureVolumePermissions` sets 0777 on the subpath, which carries
  most pairs. Two apps running as different users can still write files the other cannot
  read, and nothing here changes that.
- **Editing a dangling mount.** The form has no app to select, so saving the mount as it
  stands turns it back into one of the app's own - which moves what it points at. Deleting
  the mount is the honest answer, and the screen says so rather than the form pretending it
  can be repaired.
- **Two writers on one database directory.** Running a second process over a live Postgres
  or SQLite data directory corrupts it. Read-only is the default for a reason; see §10.

## 9. The dashboard

In *Settings > Persistent Storage*, adding a mount gains a source choice: **This app** (the
default, unchanged) or **Another app**. Choosing another app:

- lists the apps of the environment, and for the chosen one lists **its mounts** by target
  path, so picking `/var/lib/postgresql` fills in both the volume and the subpath. Nobody
  types a path.
- shows a read-only switch, on by default.
- states plainly what is granted: *"This app will be able to read the files of `postgres`."*
  With the switch off, *"read and change"*, in the warning colour.
- when the chosen app's category is `database`, turning the switch off asks for an explicit
  confirmation that says what running two writers over a data directory does.

A mount that came back `dangling` is shown with the stored path and a note that the app that
owned it no longer exists.

## 10. Templates

No new syntax for the app being named: `parameters` already has `type: app`, which
`postgres-replica` uses to pick a primary. A template can ask for an app and mount its data
(§10a covers the other direction, where the app is a dependency of the same request):

```yaml
parameters:
  - name: targetApp
    title: App to manage
    type: app

app:
  deployment:
    storage:
      mounts:
        /srv/data:
          type: volume
          source: "${{ params.dataVolume }}"
          sourceApp: {app: "${{ params.targetApp }}", write: true}
```

A `type: app` parameter stores the app's **key**, not its id, so the template writes
`sourceApp.app` and provisioning resolves it to the id before the mount is built - the same
resolution the dependency bindings already do. `specmodel/buildable.go` allows `sourceApp`
on a volume mount, with `app` and `write` as the only fields. Provisioning gates it the way capabilities are gated: the creation dialog lists
what the template will reach, in the same block that lists capabilities, and the same
pre-provision check refuses it without Write on the named app.

## 10a. A dependency that needs the main app's storage

A template's dependencies are other templates - `postgres`, `mysql` - and a dependency
template cannot name the app it was created for, because it does not know one exists. So
the mount is declared where the relationship is: in the owner template's `dependencies`
entry.

```yaml
dependencies:
  - name: files
    title: File manager
    template: filebrowser
    mounts:
      - target: /srv/data                     # where in the dependency's container
        volume: "${{ params.dataVolume }}"    # a volume parameter of this template
        subpath: ""                           # optional, below the owner's directory
        write: true                           # absent means read-only, as everywhere else
```

Nothing about the ordering gets in the way. `planApps` allocates every app's id before any
of them exists, "so that each binding can name the others before any exists" - the same
mechanism the dependency env-var bindings already rely on. The prefix a mount needs is built
from keys, not ids, and the main app's key is decided at render time too. So provisioning
passes the planned owner to `BuildAppMounts` as `OwnerApp` and the dependency comes up with
the directory already mounted, even though it is provisioned before the app that owns it.

**No permission is checked for this one.** The owner is an app of the same request, created
by the same person in the same click; there is no third party whose data is being reached.
The check in §4 is for the other case - a `type: app` parameter naming an app that was
already there:

```yaml
parameters:
  - name: targetApp
    title: App to manage
    type: app

app:
  deployment:
    storage:
      mounts:
        /srv/data:
          type: volume
          source: "${{ params.dataVolume }}"
          sourceApp: {app: "${{ params.targetApp }}", write: true}
```

That one is gated exactly as capabilities are: the creation dialog lists what the template
will reach, in the same block that lists capabilities, and the pre-provision check refuses it
without Write on the named app.

## 11. Security

The grant is the target app's data, in full, for as long as the mount exists. Three things
keep that honest:

1. It is checked against the target app, not against a module, so it cannot be obtained by
   having broad rights somewhere else.
2. It is visible from both sides: the storage screen of the *borrowing* app names the owner,
   and the storage screen of the *owning* app lists the apps that mount it. A grant nobody
   can see from the owner's side is a grant nobody will ever revoke.

   The owner's list is built by reading the services of the environment's other apps, since
   a mount lives in the service spec and nowhere else. That is one inspect per app, on a
   settings screen rather than a hot path, and an app whose service cannot be read is left
   out instead of failing the screen.
3. It is recorded: the audit entry for the storage update names the target app and whether
   write was asked for.

Read-only mounts still expose secrets that live in files - a database's data directory
contains everything in the database.

## 12. Testing

Unit, in `volumeserviceimpl`:

- a mount naming another app resolves to that app's prefix, for a project-scoped and for an
  env-scoped volume;
- naming the app itself equals naming nothing;
- naming an app in another environment is refused;
- `write: false` produces `ReadOnly: true`, and the top-level `readOnly` does not override it;
- the volume scope check still refuses a volume the caller may not mount, even when the
  target app may.

Unit, in `app_storage_remove`:

- an app with one own mount and one foreign mount deletes exactly one directory;
- a foreign mount whose owner was deleted is still not deleted by the borrower;
- a mount whose subpath equals the app's own prefix plus a deeper path is still deleted
  (the existing behaviour must not regress).

Round trip, in `appsettingsdto`:

- a foreign bind mount, transformed back, reports the owning app by name;
- an orphaned one reports `dangling`.

`docs/openapi/swagger.json` is generated and committed, so a DTO change means
`make gen-swag` in the same piece of work, and the dashboard's API validator moves with it.

Two things this had to learn the hard way, both worth keeping:

- `appScopePrefix` is reached from the deletion path, which loads the app row **without** its
  project and environment relations. It reads the environment's key out of `ProjectEnvID`
  for that reason, and answers with an empty prefix - nothing owned, nothing deleted - when
  even that is unavailable. A test covers the app loaded bare.
- On Docker Desktop for macOS the daemon rewrites a bind's source to `/host_mnt/<path>`, so
  reading a mount back to its volume finds no match and an app's storage is neither
  described nor deleted. This predates shared mounts - it is how `RemoveAppStorage` already
  behaved there - and it does not happen on Linux, which is what HivePaaS runs on.

End to end, by hand on the dev cluster: a file browser mounting a Postgres app's data
directory read-only, then deleting the file browser and confirming with `ls` on the host
that the Postgres directory is untouched.

## 13. Later

- `pathScope` for the whole-volume case a backup agent would want. This field is a special
  case of it: "the prefix is not always mine".
- Revocation: a mount that stays after the person who made it loses Write on the target app.
  Doing it properly means deciding who is asked - the owner, an admin, or a periodic
  re-check - which is more than this change should carry.
- Cross-environment sharing, for a staging app reading production data - which is a policy
  question before it is a technical one.

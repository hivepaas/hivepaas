# Backup Snapshots: Seeing What the Repositories Hold

Backup repositories exist at the global, project and env scopes. Data backup jobs
fill them, and HivePaaS keeps a record of every snapshot. Nothing shows those
records, and nothing lets a person delete a snapshot. This design adds a
**Backup Snapshots** view, which lists the snapshots of every repository a scope
sees, filtered by repository, app, tag and time.

It is the first of two pieces. The second is restore, a spec of its own that
follows at once. It adds a Restore entry to this view's context menu.

---

## Decisions

1. **Scopes.** There is one view at each scope that has backup repositories - global, project, env - and one
   at the app scope, which has none of its own but owns the data backups.
2. **Records, not live reads.** The view lists the snapshot records HivePaaS
   keeps in the database.
   - They are kept in step with each repository by the data backup job after
     every snapshot, by `backup-repo-cleanup`, and by the repository's Sync
     action.
   - Reading kopia on every visit would reach S3 or an agent per repository,
     slowly and fallibly.
3. **Every snapshot has an owner, and a person sees a snapshot when they may
   open its owner's screens.** One repository can hold the data of several
   projects; the view never shows more than the screens it links to would.
4. **Restore is not here.** The context menu has no Restore entry until the
   restore spec.

## 1. Who sees a snapshot

**The owner:**

- **A snapshot tagged `hivepaas.app:<app id>`** - every data backup snapshot is
  - is owned by that app. It is visible to a person who may read the app's env.
  - That is the question every app screen asks, through `permission.Visibility`
    (`AllowsProjectEnv`). An env's grant comes from its project's, and a project's
    from the module's; an owner sees everything.
- **A snapshot with no app tag** - a system backup, or one taken outside
  HivePaaS and synced in - is owned by its repository. It is visible to a person
  who may read the repository's scope:
  - the settings module, for a global repository;
  - the project, or the env, for a repository there.
- **A snapshot tagged with an app that is deleted** is owned by its repository.
  It shows the app as "deleted app `<id>`".

**Each view takes the snapshots in its reach, then keeps those its viewer may
see:**

| View | Snapshots of the repositories | Kept when the owner is |
|---|---|---|
| Global | at the global scope | anything |
| Project | of the project and the global scope | an app of the project, or a repository of the project or one of its envs |
| Env | of the env, its project and the global scope | an app of the env, or a repository of the env |
| App | the app sees - its env, its project, the global scope | the app |

- A project's view never shows another project's app, even when both back up
  into one global repository.
- A person with a grant on one env of a project sees only that env's snapshots
  in the project's view.
- In the global view, a person with settings access but no access to a
  project does not see that project's apps' snapshots.

## 2. API

These follow the routes of `backup-repos` at each scope:

| Scope | List | Get | Delete |
|---|---|---|---|
| Global | `GET /settings/backup-snapshots` | `GET …/:itemID` | `DELETE …/:itemID` |
| Project | `GET /projects/:projectID/backup-snapshots` | ″ | ″ |
| Env | `GET /projects/:projectID/:env/backup-snapshots` | ″ | ″ |
| App | `GET /projects/:projectID/:env/apps/:appID/backup-snapshots` | ″ | ″ |

**The list:**

- **Filters:**
  - `repo`, several;
  - `app`, several, not at the app scope;
  - `tag` as `key:value`, several, all of which must match;
  - `fromDate` and `toDate` (`YYYY-MM-DD`, both included) on the time the
    snapshot was taken, as the tasks list names them;
  - `search` on the short ID or the description;
  - paging, newest first.
- **A row** holds:
  - the record's ID and the snapshot's short ID;
  - the time, the size, the description and the tags;
  - the repository (ID, name, scope, status);
  - the app (ID, name, env) and the job (ID, name), read from the
    `hivepaas.app` and `hivepaas.job` tags;
  - the source kind, `command` or `volume`.
- **The source kind** comes from a new tag, `hivepaas.source:command|volume`, which the
  data backup job adds to every snapshot it takes from now on. For an older
  snapshot it comes from its job, when the job still exists, and is empty
  otherwise.
- **The repositories:** the response also names each repository the view
  reaches, for the filter and for the links to their Sync action. When a
  repository was last synced is not recorded anywhere today, and is not added.

**Get** holds all of a row's fields, and:
- the full snapshot ID;
- the paths and the hostname;
- the run that took it, from the `hivepaas.run` tag.

**Delete:**
1. It deletes the snapshot in the repository (`DeleteSnapshot`; through the agent
   for a repository on a volume).
2. Then it deletes the record and its tags.

It needs delete access on the scope, and on the snapshot's owner as §1 reads it,
and an active repository. kopia's "no snapshots matched" answer is the engine's
`ErrSnapshotNotFound`. A snapshot already gone from the repository counts as deleted:
the record goes.

**Errors:** a failed delete says why, with kopia's words, as every failed kopia
command now does. The record stays, for another try.

## 3. Dashboard

**Where:**

- **Global:** Settings › Automation, **Backup Snapshots**, under Backup Repos.
- **Project and env:** the project's sidebar, **Backup Snapshots**, under
  Backup Repos. Its env selector is Backup Repos': All is the project's view,
  and an env is that env's view.
- **App:** the app's Configuration menu, **Backup Snapshots**, after Scheduled
  Jobs.

**One table component serves all four,** as `BackupRepoTable` serves settings and
projects.

- **Columns:**
  - the time taken and the short ID;
  - the repository;
  - the app, not at the app scope;
  - the job;
  - the source (Command or Volume);
  - the size and the tags.
- **Order:** newest first, paged.
- **The filter bar,** laid out as Operations › Tasks' and the scheduled jobs':
  - a repository, one at a time - the API takes several;
  - an app, one at a time, at the project and env views. The global view has
    no app list to pick from; a tag `hivepaas.app:<id>` narrows it to an app;
  - tags `key:value`, several, each added with Enter;
  - a date range;
  - a search box.
  A link in sets them through the URL: `repo`, `app` and `tag`.
- **Above the table:** a note that the list is what the repositories held when
  they were last synced, with a link to each repository's page and its Sync
  action.
- **Each row's context menu:**
  - **View details:** a drawer with the full ID, paths, hostname, description,
    tags, and a link to the run that took the snapshot;
  - **Copy ID;**
  - **Delete:** confirmed by typing the short ID; shown to those who may write on
    the owner.

**Links in:**

- A data backup's run page links its snapshot ID to the app's view, filtered on
  the run.
- A data backup job's page links "Snapshots" to the app's view, filtered on the
  job.

## 4. Testing

- **Go:**
  - who sees what: an app-tagged snapshot and an untagged one, a person with a
    grant on part of a project, an owner, a deleted app;
  - each view's reach: a project does not see another project's apps through a
    shared global repository, an env sees its env only, an app itself only;
  - the filters: repository, app, several tags at once, a time range, the
    search, the paging;
  - who may delete, and deleting a snapshot already gone from the repository;
  - the new `hivepaas.source` tag on a data backup's snapshots.
- **Dashboard:**
  - typecheck, lint and prettier;
  - in a browser: the four views, the filters, the details drawer, delete, and the links in from a
    run and a job.
- **Live, on the Linux server:**
  - what a person with a grant on one env sees;
  - deleting a snapshot in an S3 repository and in one on a volume.

## Not in this version

- Restore (the next spec).
- Downloading a snapshot's files.
- Browsing the files inside a snapshot.
- Reading a repository live from the view; the records are what it shows.

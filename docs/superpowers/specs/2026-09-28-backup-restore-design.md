# Backup Restore: A Snapshot Back Into an App

The Backup Snapshots view lists what the backup repositories hold, at every
scope, and lets a person delete a snapshot. Nothing puts a snapshot's data back.
This design adds **Restore** to the view's context menu. It takes a snapshot into
the app that took it, or into another app.

It builds on the kopia engine's `RestoreStream` and `RestoreDirectory`, which
nothing calls yet. It also builds on the repository server (`OpenRepoServer`),
which serves a repository on a volume to a node that does not hold it.

---

## Decisions

1. **Into the app that took it, or into another.** Another app is how a backup
   is tried before it overwrites live data. It is also the only way back for a
   snapshot whose app is deleted, or one taken outside HivePaaS.
2. **A command snapshot goes back through a command.** The snapshot's file
   goes to the stdin of a restore command that runs in the target app, such as
   `psql`. The data backup job gains that command, `restoreCommand`, next to
   the backup command.
3. **A volume snapshot goes back into a directory.** It is restored into the
   target app's own directory in a volume, whole or one path of it, and with a
   choice of how:
   - **Replace:** the directory as it is now is moved aside, and the snapshot
     is restored into an empty one. The result is the state at backup time,
     and the old directory is kept for going back.
   - **Overwrite:** the snapshot's files are written over what is there. Files
     made after the backup stay. No extra space is needed.

   The app is stopped for the time of the restore unless the person says
   otherwise. Replace always stops it.
4. **A restore is a task.** It has a log, it can be canceled, and it shows in
   Operations › Tasks and in the target app's Tasks.
5. **Kopia runs where the data is received.** The restore follows the same
   routing a backup does.

## 1. How a restore runs

`POST …/backup-snapshots/:itemID/restore` checks the request (§2) and records a
task, `task:backup-restore`. It answers with the task's ID, and the dashboard
opens the task's page.

**A command snapshot:**

1. The command is the request's. The dialog fills it in from the job's
   `restoreCommand`.
2. `kopia show <snapshot>/<file>` reads the file.
   - For a repository on cloud storage, kopia reads it directly, from the
     backend.
   - For a repository on a volume, kopia reads it through the repository
     server, as a stream backup does.
3. The file goes to the stdin of the command, run in one running container of
   the target app. The app is not stopped: the command needs it running.
4. A command that exits non-zero fails the task, with the end of its output.

**A volume snapshot:**

1. **The target directory** is the target app's own directory in the chosen
   volume, plus the subpath, plus the snapshot path: a directory of the
   snapshot goes back to the same place under the target. It is created if it
   is not there. It is resolved as the data backup resolves its
   source, and with the same node:
   - a pinned volume is on its node;
   - a volume on all nodes is reached on the node HivePaaS runs on.
2. **Stop the app,** when asked: set it not running, and wait until none of
   its containers run.
3. **Replace:** move the directory to `<dir>.before-restore-<YYYYMMDD-HHMMSS>`
   (UTC) and make it again, empty. Both steps run as commands of the agent on the
   volume's node.
4. **Restore:** run `kopia snapshot restore <snapshot>[/<snapshotPath>] <dir>`
   on the volume's node.
   - Kopia reads the repository directly when that node reaches it: a
     repository on cloud storage, one on a volume of the same node, or one on
     a volume on all nodes.
   - Otherwise it reads through the repository server.
5. **On failure or cancel, with Replace:** remove what was restored, and move
   the old directory back.
6. **Start the app again** if step 2 stopped it. This step runs whatever
   happened before it.

**Logs.** Every step writes a line to the task's log:
- "Stopping app1";
- "Moved /…/app1 to /…/app1.before-restore-20260928-153000";
- "Restoring k1a2b3/uploads into /…/app1/uploads";
- "Starting app1".

**Cancel.** Canceling the task ends kopia or the command, and the steps of a
failure run.

**One at a time.** One restore runs into an app at a time. A second one is
refused while the first has not ended. The check reads the tasks without a
lock: two requests in the same instant can both pass.

**Once.** A restore's task never retries by itself: what a failed one left is
for a person to look at.

**Audit.** A restore is recorded against the target app, as an app update of
the section `backup-restore`, with the snapshot, the repository and the mode.

## 2. Data and API

**The data backup job:**

```go
type SchedJobDataBackup struct {
    // ... as today
    // RestoreCommand reads a backup on its stdin and loads it into the app, e.g.
    // `psql -U $POSTGRES_USER $POSTGRES_DB`. Optional; source: command only.
    RestoreCommand *CommandTemplate `json:"restoreCommand,omitempty"`
}
```

It is refused on a job whose source is a volume.

**The endpoints** follow those of `backup-snapshots` at each scope:

| Scope | Restore | Entries |
|---|---|---|
| Global | `POST /settings/backup-snapshots/:itemID/restore` | `GET /settings/backup-snapshots/:itemID/entries` |
| Project | `POST /projects/:projectID/backup-snapshots/:itemID/restore` | ″ |
| Env | `POST /projects/:projectID/:env/backup-snapshots/:itemID/restore` | ″ |
| App | `POST /projects/:projectID/:env/apps/:appID/backup-snapshots/:itemID/restore` | ″ |

**The request:**

```jsonc
{
  "targetApp": { "id": "…" },  // required
  // a command snapshot:
  "command": { /* CommandTemplate */ },
  // a volume snapshot:
  "volume": { "id": "…" },     // a volume the target app mounts as its own directory
  "subpath": "data",           // inside what the app sees of the volume; "" for all of it
  "snapshotPath": "uploads",   // inside the snapshot; "" for all of it
  "stopApp": true,
  "mode": "replace"            // "replace" | "overwrite"
}
```

The response is `{ "taskId": "…" }`.

**Checks, before the task is recorded:**

- **The snapshot** is in the view's reach, and the person may see it, by the
  owner rule of the Backup Snapshots spec.
- **The target app:** the person may write on it. It may be in another project,
  when they may write there.
- **The repository** is active.
- **A command snapshot:**
  - `command` is required;
  - `volume`, `subpath`, `snapshotPath`, `stopApp` and `mode` are empty.
- **A volume snapshot:**
  - `volume` is a volume the target app mounts as its own directory, by the
    data backup's rule for its source;
  - `subpath` and `snapshotPath` are relative, with no `..`;
  - `mode` is required;
  - `mode: replace` requires `stopApp: true`.
- **A snapshot with no source kind** - taken outside HivePaaS, or before the
  `hivepaas.source` tag - takes its kind from its content: a root holding one
  file is a command snapshot, anything else a volume snapshot.
- **One at a time:** no restore into the target app has yet to end.

**The task:**
- **Type** `task:backup-restore`, with the target app's scope. Its `targetId`
  is the snapshot record.
- **Its arguments** hold the request, the repository, and the full snapshot ID,
  as they were when it was recorded.
- **A snapshot deleted** before the task runs fails the task, saying so.
- **Where it shows:** it is added to the task types a project and an app list,
  so it shows in the target app's Tasks, the project's, and Operations ›
  Tasks.

**Entries.** `GET …/entries?path=` lists what is in the snapshot at `path`
(`kopia ls`): name, kind (file or directory), and size, directories first. It serves the dialog's
picker for `snapshotPath`. It follows the Get endpoint's checks.

**What a restore needs** comes with every snapshot the view lists:
- its job's file name, restore command (an inline script only), source volume
  and path;
- its app's project.

A restore's task gives its snapshot, repository, path and mode, for its page.

**Errors** say why, with kopia's words, as every failed kopia command does. A
failed restore leaves the target as it was:
- with Replace, the old directory is back in place;
- with Overwrite, what kopia wrote before it failed stays, and the log says so.

## 3. Dashboard

**Restore in the context menu** of every row of the Backup Snapshots views, and
in the details drawer.
- It is offered to a person who may write on the Project module; the server
  checks write on the app it goes into.
- For a snapshot whose repository is not active, it is disabled, with a tooltip
  saying why.

**The Restore drawer:**

1. **The snapshot:** its time, short ID, repository, app and job, source kind,
   and size.
2. **Target app:**
   - The default is the snapshot's app, when it exists and the person may write
     on it.
   - Another can be picked among the apps the view reaches: at the global view
     by project, then app, each app with its env; at the app view, the app
     itself.
   - When the target is not the snapshot's app, a line says: "The data of
     `<source app>` is loaded into `<target app>`."
3. **A command snapshot:**
   - The restore command, in the job form's command editor. It is filled in
     with the job's `restoreCommand`.
   - When the job has none, or is gone, the field is empty, with an example:
     `psql -U $POSTGRES_USER $POSTGRES_DB`.
   - A note: the command runs in a container of the target app, and gets
     `<file name>` on its stdin.
4. **A volume snapshot:**
   - **Target volume:** one of the volumes the target app mounts as its own
     directory, and a subpath. The defaults are the job's source volume and
     subpath, when restoring into the job's app.
   - **What to restore:** a tree of the snapshot (the Entries endpoint) to pick
     one directory from, or all of it.
   - **How:** *Replace* (the default) or *Overwrite*, each with a line saying
     what it does.
   - **Stop the app:** checked by default. It is checked and locked under
     Replace, with the reason.
   - **Space:** the size of what is restored. Under Replace, a line adds:
     "This much free space is needed; the directory as it is now is kept as
     `….before-restore-…`."
   - **Shared directory:** the apps given a directory of the target app's
     storage, if any, with a warning that they are not stopped.
5. **Confirmation:** typing the target app's name enables **Restore**.

After Restore, the dashboard opens the new task's page.

**The data backup job form** gains *Restore command* under the backup command,
for a command source only, with the same example.

**Links:**
- A restore task's page shows its snapshot, linked to the Backup Snapshots view
  filtered on its repository.

## 4. Testing

**Go, the API and its checks:**
- A person who sees the snapshot but may not write on the target app is
  refused. One who may write on an app of another project is not.
- A command snapshot without `command`, or with a volume field, is refused.
- A volume snapshot:
  - with a volume the target app does not mount as its own directory is
    refused;
  - with a `subpath` or `snapshotPath` holding `..` or starting with `/` is
    refused;
  - with `mode: replace` and `stopApp: false` is refused.
- A snapshot with no source kind: one file is a command snapshot, anything else
  a volume snapshot.
- A second restore into an app with one not ended is refused.
- `restoreCommand` on a job whose source is a volume is refused.

**Go, the task,** with a fake executor, agent and docker:
- **A command snapshot:**
  - `kopia show` on the file's path;
  - the stream reaches the command's stdin in the container;
  - a command exiting non-zero fails the task with the end of its output.
- **A volume snapshot with Replace:** the steps in order - stop, move, restore,
  start.
- **Failure and cancel:**
  - a failed restore puts the old directory back and starts the app;
  - so does a cancel.
- **Overwrite without stopping:** no stop, no move.
- **`snapshotPath`:** `kopia snapshot restore <id>/uploads <dir>`.
- **Routing:** kopia runs on the target volume's node; it reads the repository
  directly when that node reaches it, and through the repository server
  otherwise.
- **Real kopia:** an integration test that backs up and restores a directory,
  and a stream, on a local filesystem repository, as the existing kopia
  integration tests do.

**Dashboard:**
- typecheck, lint and prettier;
- in a browser:
  - the drawer for a command and a volume snapshot;
  - another target app;
  - the snapshot tree;
  - Stop locked under Replace;
  - the confirmation;
  - the task page after Restore.

**Live, on the Linux server:**
- `db.sql` restored into a Postgres app;
- a directory restored with Replace, then going back to `….before-restore-…`;
- a restore from a volume repository on another node, through the repository
  server;
- canceling a restore.

## Found on the way

Two faults of what restore stands on, fixed with it:
- An engine refused a storage that is only a repository server, so every
  backup through a server failed.
- The agent never closed a remote exec's stdin, so a command reading to the end
  of its input, such as `psql`, never finished on another node.

## Not in this version

- Downloading a snapshot's files.
- Ready-made command templates for Postgres, MySQL and others.
- A backup taken automatically before a restore.
- Restoring a command snapshot anywhere but through a command in an app.
- Cleaning up old `….before-restore-…` directories. The person deletes them.
- NFS and other driver-mounted volumes: they have no host path, and are parked
  on their own.

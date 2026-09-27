# Data Backup Job: An App's Data Into a Backup Repository

HivePaaS has backup repositories (kopia, on cloud storage or on a volume) and a
job that cleans them, but nothing puts an app's data into them. This design
adds a scheduled job type, **data backup**, that takes a snapshot of an app's
data - the output of a command such as `pg_dump`, or files in a volume the app
mounts - into a backup repository.

It is the first of two pieces. The second, a kopia repository server started
for the time of a backup or a restore, opens the combinations this one refuses
(§1) and is the ground restore will stand on; it has its own spec.

---

## Decisions

1. **A job type, `data-backup`,** on the `sched-job` setting: schedule,
   run-now, retry, timeout, notifications, triggers and job sequences come with
   it. `pre-deploy` with "Deploy waits for this job" backs a database up before
   every deploy.
2. **An app's job.** The command runs in the job's app; the volume is one the
   app mounts, read as the app sees it. Who can change the app can back its
   data up, and nothing else.
3. **The engine as it is.** The kopia engine already has `BackupStream` and
   `BackupDirectory`; nothing calls them yet. Where kopia runs follows from the
   repository and the source; the combinations it cannot reach are refused
   when the job is saved, until the repository server of the next spec.
4. **A failed command leaves no snapshot.** kopia may commit what it read
   before the command failed; the job deletes it.

## 1. Data model

```go
// base
SchedJobTypeDataBackup SchedJobType = "data-backup"

type SchedJobDataBackupSource string

const (
    SchedJobDataBackupSourceCommand SchedJobDataBackupSource = "command"
    SchedJobDataBackupSourceVolume  SchedJobDataBackupSource = "volume"
)

// entity.SchedJob has
DataBackup *SchedJobDataBackup `json:"dataBackup,omitempty"`

type SchedJobDataBackup struct {
    Source base.SchedJobDataBackupSource `json:"source"`
    // command: runs in the job's app (SchedJob.App); its stdout is the backup,
    // the file SourceFileName of the snapshot.
    SourceCommand  *CommandTemplate `json:"sourceCommand,omitempty"`
    SourceFileName string           `json:"sourceFileName,omitempty"`
    // volume: a volume the app mounts, and a path inside what the app sees of
    // it; "" for all of it.
    SourceVolume        ObjectID `json:"sourceVolume,omitzero"`
    SourceVolumeSubpath string   `json:"sourceVolumeSubpath,omitempty"`
    TargetRepository ObjectID `json:"targetRepository"`
    // Tags go onto every snapshot the job takes.
    Tags map[string]string `json:"tags,omitempty"`
}
```

**Rules**, checked when a job is created or updated:

- `data-backup` lives at the app scope only, and `SchedJob.App` is that app.
- `source: command`: `sourceCommand` and `sourceFileName` required; the file
  name is letters, digits, `.`, `-` and `_`, no `/`; the volume fields empty.
- `source: volume`: `sourceVolume` required, a volume the app mounts;
  `sourceVolumeSubpath` relative - no leading `/`, no `..` - and read from the
  root of what the app sees of the volume: the mount's own subpath, when the
  app mounts a part of a shared volume, is added. A job never reads another
  app's part of a volume.
- `targetRepository` exists, in the job's scope or inherited from its project
  or the global scope, and is active.
- **Refused, with the reason, until the repository server (next spec):**
  `source: command` into a repository on a volume - the agent's command
  protocol carries no stdin; `source: volume` on another node than the
  repository's volume.
- `GetRefObjectIDs` includes the repository and the source volume: deleting
  either while a job uses it goes through the setting-in-use check, and export
  and import map them.
- A `data-backup` job is an app's job like a `container-command` one: a step of
  a job sequence, and a job with triggers.

## 2. Running

`runJob` gains `data-backup`, handled by a new `databackupservice`. Building the
engine - storage, password, where its commands run - stays in
`backupreposervice`, which gains `BackupStream` and `BackupDirectory`, the
latter told the node to run on.

**Command.** The command runs in the app's container, without a TTY; its
stdout is piped into `kopia snapshot` running in the backend, as the file
`sourceFileName`; its stderr goes to the task's log. When the command exits
non-zero the job deletes the snapshot kopia made of what it read
(`DeleteSnapshot`) and fails.

**Volume.** The directory on the host is the volume's host path
(`resolveVolumeHostPath`), the mount's subpath and `sourceVolumeSubpath`. kopia
runs through the agent on the volume's node: for a cloud repository it
connects to the repository there, with the repository's own config file, as
today's engine does; for a volume repository that node is the repository's
own. The snapshot is of that directory.

**After a snapshot:** tags `hivepaas.job=<job id>`, `hivepaas.app=<app id>`
and the job's own; description `<job name> (run <task id>)`.
`SyncRepoSnapshots` for the repository, so its list has the new one. The task's
output records `snapshotId` and `sizeBytes`; the log, the size and the time it
took.

**Errors** fail the run with what went wrong: the repository gone or disabled,
the volume no longer mounted by the app, its node unreachable, the command or
kopia failing. Retry and timeout are the job's. Two runs into one repository
at once are left to run: kopia is made for several clients. Retention is the
repository's, applied by the `backup-repo-cleanup` job.

**Notes.** The repository's password and storage credentials reach kopia on a
node through the environment, as initializing a repository does today. A
volume the app writes to while it is read may be caught mid-write; a job
sequence whose first step stops or flushes the app, then the backup, avoids
it. The form and this spec say so.

## 3. API and dashboard

**API.** No new endpoint. A job's request and response at the app scope carry
`dataBackup: {source, sourceCommand, sourceFileName, sourceVolume {id},
sourceVolumeSubpath, targetRepository {id}, tags}`; the response names the
volume and the repository. The rules of §1 are the request's validation.
`make gen-swag`. A `sched-job-exec` task of a data backup has
`dataBackup {snapshotId, sizeBytes}` in the task API. MCP describes
`dataBackup` as made in the dashboard.

**Dashboard:**

- **App scheduled jobs:** `+ New Data Backup` beside `+ New Job Sequence`, a
  `DataBackupForm` of its own; the edit route opens the form a job's type
  needs. The form: name; **Source** tabs - **Command** (the command editor of
  the job form, and File name) or **Volume** (the app's mounted volumes, and a
  subpath); **Repository** (the backup repositories the app sees), a
  combination §1 refuses said so at once; tags; and the blocks the job forms
  share - schedule with No schedule, priority, timeout, retry, triggers,
  notification - and a note on reading a volume that is being written.
- **The list:** a `Backup · <repository>` tag.
- **A run's page:** the snapshot's ID and size. The repository's snapshot list
  has it, synced by the job.

## 4. Testing

- **Go:** the rules of §1, the refused combinations and subpaths with `..` or a
  leading `/`; the host path from the volume, the mount's subpath and the
  job's; the command run with a fake engine - a snapshot synced and recorded,
  a failed command's snapshot deleted and the run failed, kopia failing; a
  volume run on the volume's node; the task's output.
- **Live, on the Linux server:** `pg_dump` of a Postgres app into an S3
  repository; a volume into an S3 repository; a volume into a volume
  repository on the same node; a command into a volume repository refused
  when saved; a failing command leaving no snapshot; a snapshot opened with
  the kopia CLI; a backup job on `pre-deploy`, the deploy waiting for it.
- **Dashboard:** typecheck, lint, prettier; in a browser: create and edit a
  data backup with each source, the refused combination said so, the list tag,
  a run's snapshot.

## Not in this version

- The kopia repository server started for a backup or a restore (next spec),
  and with it a command into a volume repository and a volume on another node.
- Restore.
- Volumes the app does not mount; backing several apps up at once (a job
  sequence does it).

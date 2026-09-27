# Data Backup Job: An App's Data Into a Backup Repository

HivePaaS has backup repositories (kopia, on cloud storage or on a volume) and a
job that cleans them, but nothing puts an app's data into them. This design
adds a scheduled job type, **data backup**, that takes a snapshot of an app's
data - the output of a command such as `pg_dump`, or files in a volume the app
mounts - into a backup repository.

It is the first of two pieces. The second, a kopia repository server started
for the time of a backup or a restore, makes the combinations this one cannot
run yet work (§1) and is the ground restore will stand on; it has its own
spec, and follows at once.

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
   repository and the source. The combinations it cannot reach yet are not
   refused when a job is saved: the repository server of the next spec, which
   follows at once, makes them work, and a check added now would be taken out
   again. Until then their runs fail, saying why.
4. **A failed command leaves no snapshot.** kopia may commit what it read
   before the command failed; the job deletes it.
5. **One kopia source a job.** Every snapshot of a job is recorded under
   `hivepaas@data-backup:/<job id>`, whichever machine took it: kopia's
   retention goes by source, and the backend's and an agent's hostnames change
   as they are redeployed. The repository's retention keeps its count per job.

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
- `source: volume`: `sourceVolume` required, a volume the app mounts as its
  own directory - the directory HivePaaS makes for the app in a volume. A
  volume mounted whole, which carries no owner, and another app's directory
  the app was given are refused. `sourceVolumeSubpath` relative - no leading
  `/`, no `..` - and read from the root of what the app sees of the volume:
  the mount's own subpath is added. A job never reads another app's part of a
  volume. The app's service is read to check it, as the storage screen reads
  it.
- `targetRepository` exists, in the job's scope or inherited from its project
  or the global scope, and is active.
- `tags`: at most 20; a key of letters, digits, `.`, `-` and `_`, up to 50,
  not starting with `hivepaas.`; a value of up to 100 characters without
  spaces. A data-backup job has no `command`, `commandOutput` or `sequence`,
  and no other type has `dataBackup`.
- **Not run yet, and not refused when saved:** `source: command` into a
  repository on a volume - the agent's command protocol carries no stdin - and
  `source: volume` on another node than the repository's volume. A run of such
  a job fails with that reason until the repository server (next spec).
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
(`DeleteSnapshot`) and fails; when kopia named none - it was stopped, or failed
too - the snapshots tagged with the run are listed and deleted. The command
starts once kopia is connected to the repository: the two share the run's
transaction, which is not safe to use from both at once. kopia stopping first
fails the run with kopia's error, not the command's broken pipe.

**Volume.** The directory on the host is the volume's host path
(`resolveVolumeHostPath`), the mount's subpath and `sourceVolumeSubpath`. kopia
runs through the agent on the volume's node: for a cloud repository it
connects to the repository there, with the repository's own config file, as
today's engine does; for a volume repository that node is the repository's
own. The snapshot is of that directory. The node is the one the volume is
pinned to; a volume pinned to none fails the run, saying so.

**After a snapshot:** tags `hivepaas.job:<job id>`, `hivepaas.app:<app id>`,
`hivepaas.run:<task id>` and the job's own; description `<job name> (run
<task id>)`; source `hivepaas@data-backup:/<job id>`. The repository's full
list is read and `SyncRepoSnapshots` run with it, so its list has the new
one; a failure there is logged and does not fail the run - the next sync
finds it. The task's output records `snapshotId` and `sizeBytes`; the log, the
size and the time it took. A step of a job sequence leaves the task's output,
the sequence's run, alone: its step outputs are `SNAPSHOT_ID` and
`SNAPSHOT_SIZE_BYTES`.

**The engine,** fixed on the way: a stream needs a source path, which
`BackupStream` did not pass (kopia answered `no snapshot sources`); `snapshot
create --json` gives no stats, so a new snapshot's size is its root entry's
sum; `BackupOptions` gains `Description` and `Source` (`--override-source`).

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
volume and the repository. The rules of §1 are the request's validation. An
app's storage settings give each mount of the app's own directory its
`volumeId`, which the form's volume picker lists.
`make gen-swag`. A `sched-job-exec` task of a data backup has
`dataBackup {snapshotId, sizeBytes}` in the task API. MCP describes
`dataBackup` as made in the dashboard.

**Dashboard:**

- **App scheduled jobs:** `+ New Data Backup` beside `+ New Job Sequence`, a
  `DataBackupForm` of its own; the edit route opens the form a job's type
  needs. The form: name; **Source** tabs - **Command** (the command editor of
  the job form, and File name) or **Volume** (the app's mounted volumes, and a
  subpath); **Repository** (the backup repositories the app sees); tags; and the blocks the job forms
  share - schedule with No schedule, priority, timeout, retry, triggers,
  notification - and a note on reading a volume that is being written.
- **The list:** a `Backup · <repository>` tag.
- **A run's page:** the snapshot's ID and size. The repository's snapshot list
  has it, synced by the job. The task API gives `sequenceRun` only for a run
  that started: a data backup's output reads as an empty one.

## 4. Testing

- **Go:** the rules of §1 and subpaths with `..` or a leading `/`; the host path from the volume, the mount's subpath and the
  job's; the command run with a fake engine - a snapshot synced and recorded,
  a failed command's snapshot deleted and the run failed, kopia failing; a
  volume run on the volume's node; the task's output.
- **Live, on the Linux server:** `pg_dump` of a Postgres app into an S3
  repository; a volume into an S3 repository; a volume into a volume
  repository on the same node; a failing command leaving no snapshot; a snapshot opened with
  the kopia CLI; a backup job on `pre-deploy`, the deploy waiting for it.
- **Dashboard:** typecheck, lint, prettier; in a browser: create and edit a
  data backup with each source, the list tag, a run's snapshot.

## Not in this version

- The kopia repository server started for a backup or a restore (next spec),
  and with it a command into a volume repository and a volume on another node.
- Restore.
- Volumes the app does not mount, or mounts whole; backing several apps up at
  once (a job sequence does it).
- A volume pinned to no node.

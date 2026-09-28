# System Backup Into a Backup Repository

The system backup job dumps HivePaaS's own database into a tar file. The file
can be compressed and encrypted. It is kept on the local disk, listed on a
**Backup Files** page, and uploaded to cloud storage when that is configured.

This design changes two things:
- **What is backed up:** the database, the configuration **spec** of the whole
  installation, or both.
- **Where it goes:** a kopia **backup repository**, as app data does, in place
  of files.

It adds a way to take a file out of any snapshot: a download. Backward
compatibility is not kept.

---

## Decisions

1. **The database, the spec, or both,** as the configuration says; one at
   least.
2. **The spec's secrets are the operator's choice:**
   - **encrypted** with a passphrase of the system backup's (the default);
   - **omitted**;
   - **plaintext.**

   The database dump holds its secrets encrypted with a data key that the app
   secret wraps, and the app secret is not in the backup. A plaintext spec would
   be the weakest part of it, so it is not the default.
3. **One snapshot a run,** holding `db.pg_dump` and `spec.tar.gz`: the two
   taken at one time, one row a run. kopia keeps what did not change once.
4. **The repository compresses and encrypts.** The system backup's own
   compression, encryption, cloud storage and local files go. So does the
   Backup Files page.
5. **Restore is not here.** A file of a snapshot can be downloaded:
   - the spec imports through the Import page;
   - the dump restores by hand with `pg_restore`.

   An automatic restore of HivePaaS's database replaces the database of the
   running system itself, and has its own spec.

## 1. The configuration

```go
type SystemBackup struct {
    Schedule    SchedJobSchedule
    // What goes in: one at least.
    IncludeDB   bool
    IncludeSpec bool
    // SpecSecrets is how the spec holds secrets: encrypted (the default),
    // omit, plaintext. SpecPassphrase is required when encrypted.
    SpecSecrets    specmodel.SecretsMode
    SpecPassphrase EncryptedField
    // TargetRepository is a backup repository at the global scope.
    TargetRepository ObjectID
    Notification     *BaseEventNotification
}
```

**Gone:**
- compression (gzip, zstd) and encryption (age);
- cloud storage, its bucket and directory;
- `backupDeletedObjects`, which nothing reads;
- the local backup directory, the `system-backup` file kind;
- the Backup Files API and page. Files left on disk by earlier runs are not
  removed; the operator deletes them.

**Checks, when the configuration is saved:**
- one of the database and the spec at least, and a repository at the global
  scope - of an enabled configuration: a disabled one may name neither;
- a passphrase when the spec's secrets are encrypted;
- a spec holding secrets, encrypted or plaintext, asks the person saving it
  for the capability to reveal secrets, recorded as a reveal. It is the
  capability alone, as a mount's is: the backup puts the secrets in a
  repository and hands the caller nothing in the clear.

**Defaults:** disabled, the database only, secrets encrypted, no repository.

The passphrase is stored encrypted, and the API masks it as it masks every
secret.

## 2. A run

The system backup task runs as it does today:

1. **A work directory** in the backend.
2. **The database:** `pg_dump -Fc` into `db.pg_dump`, the custom format that
   `pg_restore` reads selectively. The `migrations` table is left out, as today.
3. **The spec:** the spec export at the global scope, with the configured
   secrets mode and passphrase, into `spec.tar.gz` - `spec.tar.gz.age` when
   encrypted, as the export names an encrypted bundle.
4. **One snapshot of the work directory:**
   - source `hivepaas@system-backup:/system`;
   - tags `hivepaas.source:system-backup` and `hivepaas.run:<task id>`;
   - description `System backup (run <task id>)`.

   kopia runs in the backend. It writes into a repository on cloud storage
   directly, and into one on a volume through the repository server, as
   `hivepaas@system-backup`. The backup repository service gains that: a
   backup of a directory of the backend's own.
5. **The repository's snapshot records** are brought up to date, as a data
   backup does, so the snapshot shows at once.
6. **The task's output** is the snapshot's ID, its size, and what it holds,
   named as a data backup's are, so the run's page reads it the same way.

**Failure:**
- A failed dump or export fails the run before anything is taken. The log
  says which step failed.
- A snapshot left half made is deleted by the run's tag.

**Where it shows:** the snapshot carries no app tag, so its repository owns it.
It shows in Settings › Backup Snapshots, to those who may read the Settings
module. The repository's retention counts its source apart from every other.

## 3. Downloading a file of a snapshot

```
GET …/backup-snapshots/:itemID/download?path=<a file in the snapshot>
```

It is served at the four scopes, as the snapshots' other endpoints are.

**Checks:**
- **Who:** the snapshot is in the view's reach for a person who may **write**
  on its owner, as for Delete and Restore. A download is the data itself: an
  app's database dump, or HivePaaS's own.
- **The repository** is active.
- **`path` is a file.** A directory is refused with a 400 that says so;
  Restore is how a directory comes back.

**The answer:**
- kopia's `show` is streamed into the response, never to disk.
- The headers:
  - `Content-Disposition: attachment; filename=<its name>`;
  - `Content-Type: application/octet-stream`;
  - `Content-Length` from the snapshot's listing.
- A repository on a volume is read through the repository server, as a
  stream restore is.
- **A failure half way** cuts the connection, and the download fails: once
  the headers are sent there is no other way to say so.

**Audit:** a download is recorded, as `backup-download`, with who, the
snapshot, the repository and the path:
- against the app, for an app's snapshot;
- against the repository otherwise.

**The system cleanup** keeps no backup files any more: its retention of them
goes, from its settings, its run and its form.

## 4. Dashboard

**Settings › Data Backup › Configuration** is a new form:
- **What to back up:** *Database*, HivePaaS's database dump; *Spec*, the
  configuration of every project, env and app, as Export writes it. One at
  least.
- **Secrets in the spec,** when the spec is included:
  - *Encrypted* (the default), *Omit*, *Plaintext*, each with a line saying
    what it means;
  - *Encrypted* asks for a passphrase, masked, with a note: a lost passphrase
    is a spec whose secrets cannot be read.
- **Backup repository:** one at the global scope, with a link to Backup
  Repos.
- **Schedule** and **notification** as today.
- A **View snapshots** link to Settings › Backup Snapshots, filtered on
  `tag=hivepaas.source:system-backup`.

**The Backup Files page goes,** with its tab, its route, its API and its data
layer.

**Settings › Data Backup › Actions** stays. A run's card shows the snapshot it
took, with its size, linked to Backup Snapshots filtered on its
`hivepaas.run` tag, as a data backup's run does.

**Backup Snapshots, at every view:**
- **The source column** reads *System* for a system backup's snapshot.
- **The details drawer** gains **Files:** the snapshot's tree, browsed as the
  Restore page browses it. Each file has its size and a **Download** button.
  - The button is shown to those who may write on the snapshot's owner.
  - A download is an authenticated request read as a blob, as the Backup
    Files page did. The button shows it is working meanwhile.
  - The blob is held in the browser's memory: a file of several gigabytes is
    beyond it. A short-lived download link is for a later spec.

## 5. Testing

**Go, the configuration:**
- neither the database nor the spec is refused;
- encrypted without a passphrase is refused;
- a repository not at the global scope, or not there, is refused;
- the passphrase is stored encrypted and read back masked.

**Go, a run,** with a fake `pg_dump`, spec export and repository service:
- the database only, the spec only, both: the directory taken holds
  `db.pg_dump`, `spec.tar.gz`, or both;
- the export is asked for at the global scope, with the secrets mode and the
  passphrase;
- the snapshot's source, tags and description; the records are updated;
- a failed dump or export fails the run with no snapshot, and says which step;
- a snapshot failing half way is deleted by the run's tag;
- a directory of the backend's own:
  - into cloud storage directly;
  - into a volume repository through the server, as `hivepaas@system-backup`.

**Go, a download:**
- a person who may only read is refused; one who may write on the owner is
  not;
- a directory is a 400; a path that is not there is not found;
- the headers, and the bytes, against real kopia on a filesystem repository;
- the audit entry.

**Go, real kopia,** as the existing integration tests do: a directory of two
files taken, listed, and one of them downloaded.

**Dashboard:**
- typecheck, lint and prettier;
- in a browser:
  - the configuration form: the two choices, the secrets mode, the
    passphrase, the repository;
  - no Backup Files page;
  - a run's snapshot link;
  - downloading `db.pg_dump` and `spec.tar.gz` from the details drawer.

**Live, on the Linux server:**
- a system backup into an S3 repository, and into a volume repository;
- the downloaded `spec.tar.gz` imported through the Import page, with its
  passphrase;
- `pg_restore --list db.pg_dump` reads the dump.

## Not in this version

- Restoring HivePaaS's database automatically.
- Downloading a directory, or a very large file through a short-lived link.
- Static files: the list of directories a system backup took is empty today,
  and goes.

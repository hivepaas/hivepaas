# Files on Volumes

A file HivePaaS keeps on its own disks - a repository cache, a scheduled job's
output, a file uploaded to an app - is an `entity.File` with storage type
`local` and a path under `config.AppPath`. That path is read with `os.*` by
whichever process handles the file, on whichever node it runs. It works today
only because the app, the worker and the database are all held to the one
node labelled `hivepaas.role=control-plane`: "local" happens to mean "that
node".

It stops working when anything else is true:
- **an agent on another node** has to read or write a file - the image build
  on a build node is the first case, a file made by a job on another node the
  next;
- **more than one node carries the control-plane label**: the app and the
  worker can land on different nodes, and a file the worker wrote is missing
  where the app looks for it;
- **the control plane moves** to another node: the files stay behind, and
  nothing in the database says where.

This design binds every such file to a **volume**. The volume, not the file,
says where the data is: on one node, or on every node. Backward compatibility
is not kept.

---

## Decisions

1. **A file on HivePaaS's disks is on a volume.** Storage type `local` becomes
   `volume`; `cloud` stays.
2. **The volume says where.** It is the `ClusterVolume` setting HivePaaS
   already has, pinned to a node or bound to a directory every node shares.
   Nothing about a file's location is derived from configuration.
3. **The system volume is implicit**: HivePaaS's data directory on the
   control-plane node, the one the app and the worker already mount. It has no
   record.
4. **Project data goes to the project's default volume**: the repository cache,
   scheduled job output and the files uploaded to an app, in a `.hivepaas/`
   directory of their own.
5. **Every read, write and removal goes through one access layer.** The system
   volume is read directly; any other volume through the agent of its node.
6. **A volume that holds files cannot be deleted.** The files are counted by a
   function of their own, so deleting them with the volume can be added later.

## 1. `entity.File`

```go
type File struct {
    // ...
    // StorageType is volume or cloud.
    StorageType base.FileStorageType
    // StorageID is the ClusterVolume setting for a file on a volume, empty for
    // the system volume; the cloud storage setting for a file in the cloud.
    StorageID string
    // Path is inside the volume, or the object key in the cloud.
    Path string
}
```

- `base.FileStorageLocal` is replaced by `base.FileStorageVolume` (`"volume"`),
  in the API, the seed data and the dashboard. The column is a free `VARCHAR`,
  so no migration changes; a database that is kept needs
  `UPDATE files SET storage_type = 'volume' WHERE storage_type = 'local'`, and
  its files are then on the system volume, where their paths already point.
- `StorageID` already exists; it names the volume as it names the cloud
  storage.
- `Path` is relative to the volume's directory. It never starts with `/` and
  never contains `..`.

## 2. Where each file goes

| File | Scope | Volume | Path in it |
|---|---|---|---|
| Repository cache | project | the project's default volume | `.hivepaas/cache/repos/<id>.<random>.<ext>` |
| Scheduled job output, not to the cloud | app | the project's default volume | `.hivepaas/job-output/<env>/<app>/<id>-<name>` |
| A file uploaded to an app (data files) | app | the project's default volume | `.hivepaas/files/<env>/<app>/<id>-<name>` |
| A file uploaded at any other scope (`tmp`) | global, user | the system volume | `files/<id>-<name>` |

- **The project's default volume** is the `ClusterVolume` a project gets when
  it is created: `default`, a bind to `project_data/<project key>`, pinned to
  the node it was created on. It is the project's setting marked `Default`.
- **`.hivepaas/` is HivePaaS's, with mode `0700`.** An app's volumes are made in
  `<project>/<env>/<app>`, so no app sees it; a volume someone binds to the
  project's own directory would, and the name says what it is.
- **A file records the volume it was written to.** When a project's default
  volume changes, its existing files stay where they are and are still found;
  new ones go to the new default.
- **The image build settings get no volume of their own.** The cache follows
  the project.

## 3. Which volumes can hold files

The rules of a backup repository on a volume (`volumeHostDir`), with one more:

| Volume | Can hold files | Reached on |
|---|---|---|
| pinned by `NodeID` | yes | that node |
| pinned by `NodeLabel`, matching one node | yes | that node |
| pinned by `NodeLabel`, matching several or none | no: which node holds the files cannot be told | - |
| a bind directory, unpinned (shared) | yes | the node that needs the file |
| NFS or docker-managed, unpinned | no, as for backup repositories | - |
| a node no longer in the cluster | no: the agent service's own error, no agent on that node | - |

When the project's default volume cannot hold files:
- **the repository cache** is skipped: the deployment logs a warning naming the
  volume and why, and checks out without a cache;
- **a scheduled job's output** fails the job with the same message;
- **an upload** is refused with it.

## 4. The access layer

In `fileservice`, used by everything that touches a file's content:

```go
// Open reads the file.
Open(ctx, db, file *entity.File) (io.ReadCloser, error)
// Create writes a file: to a temporary name first, renamed into place when
// the writer is closed without error, so no half-written file is ever found.
// Abort discards what was written instead.
Create(ctx, db, file *entity.File) (FileWriter, error)
// Remove removes the file; a file already gone is not an error.
Remove(ctx, db, file *entity.File) error
// Stat is the file's size, or not found.
Stat(ctx, db, file *entity.File) (int64, error)
// CountOnVolume counts the files a volume holds.
CountOnVolume(ctx, db, volumeID string) (int, error)
// ProjectVolume is the project's default volume.
ProjectVolume(ctx, db, projectID string) (*entity.Setting, error)
```

- **`cloud`**: the S3 code that exists today.
- **The system volume**: directly under `config.AppPath`, as today.
- **A `ClusterVolume`**: `volumeservice.ResolveHostDir` resolves its directory
  on the host and its node - moved from `backupreposervice`, so backups and
  files resolve volumes the same way. The layer then calls the agent of that
  node, **even when it is this node**: the app's container mounts the data
  directories, not every volume. One path, the same everywhere.
- **The repository cache passes through the checkout's temporary directory**:
  the archiver works on paths, so the archive is compressed there and copied to
  the volume, and copied back to be unpacked.
- **The callers change to it**: `fileuc` (upload, download, delete),
  `fileservice` (upload, delete), the repository cache (load, save, cleanup,
  size), the scheduled job's output (write, remove on failure). No code outside
  the layer joins a file's path with `AppPath` any more.

## 5. The agent's `FileService`

```proto
service FileService {
  rpc FileRead(FileReq) returns (stream FileChunk);
  rpc FileWrite(stream FileWriteReq) returns (FileWriteResp);
  rpc FileRemove(FileReq) returns (FileRemoveResp);
  rpc FileStat(FileReq) returns (FileStatResp);
}
message FileReq {
  string root = 1;  // the volume's directory on the host
  string path = 2;  // relative to root
}
```

- The agent sees the host at `/host`, and works in `/host<root>/<path>`.
- **It refuses a path that leaves the root**: an absolute path, a `..`, or a
  symbolic link that resolves outside it (the path is resolved with
  `EvalSymlinks` and must stay under the resolved root).
- **A write goes to `<path>.tmp-<random>`**, then is renamed onto `<path>` when
  the stream ends without error; a failed or cancelled write removes the
  temporary file.
- The agent already runs any command it is asked to (`ExecuteCommand`), so this
  service widens nothing the agent can do; the path checks keep a mistake in
  HivePaaS from reaching outside a volume.

## 6. Deleting a volume

- **Refused while it holds files**: `ERR_VOLUME_HAS_FILES`, with the count.
  Removing the files, or the app or project they belong to, frees it.
- `CountOnVolume` is its own function, so an option to delete the files with
  the volume is only a loop over them through `Remove`.

## 7. Dashboard

- `AppDataFileStorageType.Local` becomes `Volume` (`"volume"`), in the upload
  and in the table's label.
- The table of an app's data files reads "HivePaaS" for a file on a volume, as
  it read "Local".

## 8. What this does not change

- **The image build** (a separate design): the app still checks out; the cache
  it loads and saves moves to the project's volume through the access layer.
- **Cloud storage**: files in the cloud are unchanged.
- **Backups**: a backup repository on a volume keeps its own engine; only the
  volume resolution is shared.

## 9. Testing

**Go, the access layer**, with a fake agent:
- a file on the system volume is read and written under `AppPath`;
- a file on a pinned volume goes to the agent of its node, with the volume's
  directory as root;
- a file on a shared bind volume goes to the agent of the current node;
- a volume pinned by a label matching two nodes, an NFS volume, a volume on a
  node that is gone: each is refused with its reason.

**Go, the agent's `FileService`**, on a real temporary directory:
- read, write, stat and remove inside the root;
- `../x`, `/etc/passwd`, and a symbolic link to outside the root are refused;
- a write that fails half way leaves no file and no temporary file.

**Go, the callers:**
- the repository cache is written to `.hivepaas/cache/repos` of the project's
  default volume, and a deployment whose default volume cannot hold files
  checks out without it;
- a job's output is written to `.hivepaas/job-output/...`, and removed when the
  job fails;
- deleting a volume holding files is refused with the count.

**Live, on the Linux server:** a deployment with the cache, a scheduled job's
output downloaded, a data file uploaded and downloaded; then the same with the
project's default volume pinned to another node.

## Later

- **`project_data` configurable, and marked shared.** Today it is set at
  install (`HP_STORAGE_PROJECT_DATA_HOST_DIR`). A flag saying it is mounted at
  the same path on every node would create new projects' default volumes as
  unpinned binds, which this design already treats as reachable from every
  node - for files, and for the apps using them, which would no longer be held
  to one node. Before the flag takes effect, HivePaaS should check it: one
  node's agent writes a marker, the others read it. Existing projects keep
  their pinned volumes until they are moved on purpose.
- **Deleting a volume with its files**, through `CountOnVolume` and `Remove`.
- **The image build on another node**, which streams the checkout to the build
  node's agent (the build context design).

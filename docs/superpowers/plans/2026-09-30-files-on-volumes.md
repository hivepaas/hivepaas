# Files on Volumes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every file HivePaaS keeps on its own disks is on a volume, and the volume says which node holds it: the repository cache, a scheduled job's output and an app's uploaded files go to the project's default volume, and are reached through one access layer.

**Architecture:**
- **Where a volume is:** `volumeservice.ResolveHostDir`, moved out of the backup repository service, gives a volume's directory on the host and its node. Backups and files resolve a volume the same way.
- **The agent:** a new `FileService` reads, writes, stats and removes a file inside a volume's directory on its node, and refuses any path that leaves it.
- **The access layer:** `fileservice` gains `Open`, `Create`, `Remove`, `Stat`, `ProjectVolume` and `CountOnVolume`. The system volume (no `StorageID`) is HivePaaS's data directory, read directly; any `ClusterVolume` goes through the agent of its node.
- **The callers:** uploads, downloads, the repository cache, the cache cleanup and the job output stop joining paths with `AppPath` and use the layer.

**Tech Stack:** Go (bun, fx, gRPC with protoc 35 and protoc-gen-go), React/TypeScript for the one dashboard change.

**Spec:** `docs/superpowers/specs/2026-09-30-files-on-volumes-design.md`

## How the patches work

Every task's code was written and verified before this plan.

**Bases:**
- The backend patches apply in order to `main` at `3af9415d`. This plan's commit changes only `docs/`.
- The dashboard patch applies to the dashboard's `main` at `0e37b0c2`.

**What was checked, on a fresh worktree, at every backend task:**
- the tests alone fail (a build error: the code they call does not exist yet);
- with the code, the task's packages pass;
- `golangci-lint run ./...` finds 0 issues, and `go run ./tools/goroutinelint .` and `go run ./tools/errcodelint` pass.

**After the last task:** `go test ./...` passes; `make gen-swag` and `go generate ./hivepaas_app/interface/agent/proto/...` change nothing; the tree equals the branch the patches were cut from.

**Dashboard:** `npx tsc --noEmit`, `npm run lint` and `npx prettier --check src` pass with the patch.

Patches live in `docs/superpowers/plans/2026-09-30-files-on-volumes/`: `be-NN-tests.patch` then `be-NN-code.patch` for backend task NN, and `dash-01.patch`.

The commands below run from the backend repo root, with `P=docs/superpowers/plans/2026-09-30-files-on-volumes` and `D=../hivepaas-dashboard`.

## Global Constraints

- **Before a backend task is done:** `go build ./...`; `golangci-lint run ./...` over the whole repo; `go run ./tools/goroutinelint .`; `go run ./tools/errcodelint`; the task's tests.
- **No import the main checkout has not vendored**, above all in tests: `testify/assert` is vendored, `testify/require` is not. After the merge, `go test ./...` runs in the main checkout too.
- **Storage types:** `volume` and `cloud`; `local` is gone. A file on a volume has `StorageID` = the `ClusterVolume` setting's id, or empty for the system volume.
- **Paths:** relative to the volume's directory, never absolute, never with `..`.
  - repository cache: `.hivepaas/cache/repos/<name>`;
  - job output: `.hivepaas/job-output/<env>/<app>/<id>-<name>`;
  - an app's uploaded file: `.hivepaas/files/<env>/<app>/<id>-<name>`;
  - any other upload: `files/<id>-<name>` on the system volume.
- **Directories HivePaaS makes for its files:** mode `0700`.
- **A location is read from the volume's record**, never computed from configuration (`project_data`).
- **Error codes:** `ERR_FILE_PATH_OUTSIDE_ROOT`, `ERR_VOLUME_CANNOT_HOLD_FILES`, `ERR_VOLUME_HAS_FILES`, each with its message in `errors.infra.en.toml`.
- **Generated code:** `file.pb.go` and `file_grpc.pb.go` come from `go generate ./hivepaas_app/interface/agent/proto/...` (protoc 35.1); they are in the code patch.
- **Git:** commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`; merge locally, never push; never `git stash`; leave the user's uncommitted work alone.

## Review Focus

1. **A file on a volume pinned to another node, live.** The agent is reached over the overlay network and a repository cache is hundreds of megabytes.
   - Pinned by: `TestAFileGoesThroughTheAgentAndBack` (Task 2, a 320 KB file over a real gRPC connection), and the live check of Task 8.
2. **Rows written before this change** have `storage_type = 'local'`. Nothing reads them as files any more.
   - Pinned by: Task 8, step 1: `UPDATE files SET storage_type = 'volume' WHERE storage_type = 'local';` on a database that is kept. Their paths stay valid: no `StorageID` is the system volume.
3. **A project whose default volume cannot hold files** (an NFS volume, a label on several nodes): the deployment must go on without a cache, and a job must fail saying why.
   - Pinned by: `TestAVolumePinnedByALabelHoldsFilesOnlyOnOneNode`, `TestAnNFSVolumeCannotHoldFiles` (Task 3); the warning in the deployment log by Task 8, live.
4. **A job that fails half way** must leave no output file: the cleanup used to be deferred with a stale error and never ran.
   - Pinned by: `TestAFailedJobLeavesNoOutputFile` (Task 6).
5. **An app whose env is not loaded on the job's request**: the output path then uses the env's id where its key should be.
   - Pinned by: Task 8, live: a scheduled job saving to a file, and the path of the file it made.

## Decisions beyond the spec

- **`Create` returns a `FileWriter`**, an `io.WriteCloser` with `Abort(err)`. The spec said `io.WriteCloser`; a caller whose source fails half way needs to discard what it wrote rather than close it into place.
- **`ProjectVolume(ctx, db, projectID)`** is a method of `fileservice`: the project's active `ClusterVolume` marked default.
- **`ResolveHostDir` is a package function** of `volumeservice` taking the docker manager, not a method: the backup repository service calls it with its own manager and keeps its tests. Its answer gains `Shared`.
- **`agentservice.NodeIDsWithLabel`** lists the ready nodes carrying a label; `GetAgentAddrForNodeLabel` is built on it. A label on no node and a label on several are both `ERR_VOLUME_CANNOT_HOLD_FILES`.
- **A node no longer in the cluster** answers with the agent service's own error ("no running agent task found on node"), not a message of the file layer's.
- **The repository cache's archive passes through the checkout's temporary directory**: it is compressed there and copied to the volume, and copied back there to be unpacked, because the archiver works on paths.
- **No migration of the `files` table**: `storage_type` is a free `VARCHAR`. The seed data changes; a database that is kept needs the `UPDATE` of Review Focus 2.
- **`deletePermanentlyIfLocal`** becomes `deletePermanentlyIfOnVolume` on the file delete request. The dashboard does not send it.
- **`config.DataPathSystemCache` and `DataPathSystemCacheRepos` go**: nothing keeps a cache in the data directory any more.
- **Two bugs the move fixes:** an app's uploaded file was written to the root of the data directory under its bare name, so two files of one name overwrote each other; and a failed job's output file was never removed.

---

## Task 1: One resolver for where a volume is

**Files:**
- Create: `hivepaas_app/service/volumeservice/host_dir.go`
- Modify: `hivepaas_app/service/backupreposervice/service.go`, `types.go`, `backupreposerviceimpl/backup.go`, `engine.go`, `engine_test.go`
- Test: `hivepaas_app/service/volumeservice/host_dir_test.go`

**Interfaces:**
- Produces:
  - `volumeservice.HostDir{Dir, NodeID, NodeLabel string; Shared bool}`;
  - `volumeservice.ResolveHostDir(ctx, dockerManager docker.Manager, setting *entity.Setting) (*HostDir, error)`;
  - `volumeservice.BindDirectory(*entity.ClusterVolume) (string, bool)`;
  - `backupreposervice.Service.VolumeHostDir` now returns `*volumeservice.HostDir`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-01-tests.patch`

Tests added: `TestResolveHostDirOfAPinnedBindVolume`, `TestResolveHostDirInspectsAVolumeWithoutADevice`, `TestResolveHostDirOfASharedBindVolumeIsOnTheCurrentNode`, `TestResolveHostDirRefusesASharedVolumeThatIsNotABind`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/service/volumeservice/`
Expected: FAIL, build error `undefined: ResolveHostDir`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-01-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/volumeservice/... ./hivepaas_app/service/backupreposervice/... ./hivepaas_app/service/databackupservice/... && golangci-lint run ./...`
Expected: PASS; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "refactor(volume): one resolver says where a volume's data is and on which node

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 2: The agent's file service

**Files:**
- Create:
  - `hivepaas_app/interface/agent/proto/file.proto`, and the generated `file.pb.go`, `file_grpc.pb.go`;
  - `hivepaas_app/usecaseagent/fileagentuc/uc.go`, `file.go`;
  - `hivepaas_app/interface/agent/server/fileservice/file.go`, `hivepaas_app/interface/agent/server/server_file.go`;
  - `hivepaas_app/interface/agent/client/fileservice/service.go`.
- Modify: `hivepaas_app/interface/agent/proto/generate.go`, `hivepaas_app/interface/agent/server/server.go`, `hivepaas_app/cmd/agent/main.go`, `hivepaas_app/registry/provides.go`, `hivepaas_app/hperrors/errors_infra.go`, `hivepaas_app/pkg/translation/messages/en/errors.infra.en.toml`
- Test: `hivepaas_app/usecaseagent/fileagentuc/file_test.go`, `hivepaas_app/interface/agent/client/fileservice/service_test.go`

**Interfaces:**
- Produces:
  - agent use case: `fileagentuc.New(logger)`, `NewOnHost(logger, hostPrefix)`; `(*UC).Open(root, path) (io.ReadCloser, int64, error)`, `Stat(root, path) (int64, error)`, `Remove(root, path) error`, `Write(ctx, root, path, io.Reader) (int64, error)`;
  - gRPC: `FileService` with `FileRead`, `FileWrite`, `FileRemove`, `FileStat`; `FileReq{root, path}`;
  - client, package `hivepaas_app/interface/agent/client/fileservice` (imported as `agentfile`): `FileServiceClient{Read, Write, Remove, Stat, Close}` and `NewFileServiceClient(agentAddr)`;
  - `hperrors.ErrFilePathOutsideRoot`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-02-tests.patch`

Tests added: `TestAFileIsWrittenReadStatedAndRemovedInsideTheRoot`, `TestAPathLeavingTheRootIsRefused`, `TestAFailedWriteLeavesNothing`, `TestAWriteReplacesTheFileWhole`, `TestAFileGoesThroughTheAgentAndBack`, `TestTheAgentRefusesAPathOutOfTheVolume`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/usecaseagent/fileagentuc/ ./hivepaas_app/interface/agent/client/fileservice/`
Expected: FAIL, build errors `undefined: UC` and `undefined: hperrors.ErrFilePathOutsideRoot`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-02-code.patch`

The patch carries the generated proto code. `go generate ./hivepaas_app/interface/agent/proto/...` must leave it unchanged.

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/usecaseagent/fileagentuc/ ./hivepaas_app/interface/agent/... ./hivepaas_app/cmd/... && golangci-lint run ./... && go run ./tools/goroutinelint . && go run ./tools/errcodelint`
Expected: PASS, the fx wiring test included; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(agent): a file service reads and writes inside a volume's directory

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 3: The access layer, and `volume` for `local`

**Files:**
- Create: `hivepaas_app/service/fileservice/fileserviceimpl/access.go`
- Modify:
  - `hivepaas_app/base/file.go`, `hivepaas_app/entity/file.go`;
  - `hivepaas_app/service/fileservice/service.go`, `types.go`, `fileserviceimpl/service.go`, `delete.go`;
  - `hivepaas_app/service/agentservice/service.go`, `agentserviceimpl/agent.go`;
  - every reference to `base.FileStorageLocal`;
  - `hivepaas_app/hperrors/errors_infra.go` and its messages;
  - `hivepaas_app/service/nodeexecservice/nodeexecserviceimpl/node_exec_test.go` (its agent double gains the new method).
- Test: `hivepaas_app/service/fileservice/fileserviceimpl/access_test.go`

**Interfaces:**
- Consumes: Task 1's `ResolveHostDir`, Task 2's `agentfile.FileServiceClient`.
- Produces:
  - `base.FileStorageVolume = "volume"`; `(*entity.File).IsOnVolume()`;
  - `fileservice.FileWriter` = `io.WriteCloser` + `Abort(err error)`;
  - `fileservice.Service`: `Open(ctx, db, file) (io.ReadCloser, error)`, `Create(ctx, db, file) (FileWriter, error)`, `Remove(ctx, db, file) error`, `Stat(ctx, db, file) (int64, error)`, `CountOnVolume(ctx, db, volumeID) (int, error)`;
  - `fileserviceimpl.New(fileRepo, settingRepo, dockerManager, agentService)`;
  - `agentservice.Service.NodeIDsWithLabel(ctx, label) ([]string, error)`;
  - `hperrors.ErrVolumeCannotHoldFiles`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-03-tests.patch`

Tests added: `TestAFileOnTheSystemVolumeIsUnderTheDataDirectory`, `TestAFileOnAPinnedVolumeGoesToItsNode`, `TestAFileOnASharedVolumeGoesToTheCurrentNode`, `TestAVolumePinnedByALabelHoldsFilesOnlyOnOneNode`, `TestAnNFSVolumeCannotHoldFiles`, `TestAFilePathOutOfItsVolumeIsRefused`, `TestAnAbortedWriteLeavesNothing`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/service/fileservice/...`
Expected: FAIL, build errors `undefined: base.FileStorageVolume` and `unknown field settingRepo in struct literal of type service`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-03-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/fileservice/... ./hivepaas_app/service/agentservice/... ./hivepaas_app/service/nodeexecservice/... ./hivepaas_app/cmd/... && golangci-lint run ./... && go run ./tools/errcodelint`
Expected: PASS; `0 issues.` `git grep -n FileStorageLocal -- hivepaas_app` finds nothing.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(files): a file on a volume is reached through one access layer

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 4: Uploads and downloads

**Files:**
- Create: `hivepaas_app/service/fileservice/placement.go`
- Modify: `hivepaas_app/service/fileservice/service.go`, `types.go`, `fileserviceimpl/access.go`, `upload.go`; `hivepaas_app/usecase/fileuc/upload.go`, `download.go`, `delete.go`, `filedto/delete.go`
- Test: `hivepaas_app/service/fileservice/fileserviceimpl/access_test.go`

**Interfaces:**
- Consumes: Task 3's layer.
- Produces:
  - `fileservice.ProjectFilesDir = ".hivepaas"`, `FilesDirUploads = "files"`, `FilesDirJobOutput = "job-output"`, `FilesDirRepoCache = "cache/repos"`;
  - `fileservice.AppFilePath(dir, envKey, appKey, name string) string`, `fileservice.ProjectFilePath(dir, name string) string`;
  - `fileservice.Service.ProjectVolume(ctx, db, projectID string) (*entity.Setting, error)`;
  - `fileservice.UploadReq.App *entity.App`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-04-tests.patch`

Tests added: `TestAFileUploadedToAnAppGoesToItsProjectVolume`, `TestAFileUploadedOutsideAProjectIsOnTheSystemVolume`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/service/fileservice/...`
Expected: FAIL, build error `unknown field App in struct literal of type ...fileservice.UploadReq`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-04-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/fileservice/... ./hivepaas_app/usecase/fileuc/... && golangci-lint run ./...`
Expected: PASS; `0 issues.` `make gen-swag` changes nothing.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(files): a file uploaded to an app goes to its project's volume

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 5: The repository cache

**Files:**
- Modify: `hivepaas_app/service/repocheckoutservice/repocheckoutserviceimpl/service.go`, `repo_cache.go`; `hivepaas_app/service/syscleanupservice/syscleanupserviceimpl/service.go`, `sys_cleanup_cache.go`
- Test: `hivepaas_app/service/repocheckoutservice/repocheckoutserviceimpl/repo_cache_test.go`

**Interfaces:**
- Consumes: Task 3's `Open`, `Create`, `Remove`; Task 4's `ProjectVolume`, `ProjectFilePath`, `FilesDirRepoCache`.
- Produces: `repocheckoutserviceimpl.New(db, fileRepo, fileService)`; `syscleanupserviceimpl.New(…, auditService, fileService, dockerManager)`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-05-tests.patch`

Tests added: `TestARepositoryCacheArchiveGoesThroughTheFileLayer`, `TestARepositoryCacheIsPlacedInTheProjectVolume`, `TestARepositoryCacheNeedsAProjectVolume`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/service/repocheckoutservice/...`
Expected: FAIL, build error `unknown field fileService in struct literal of type service`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-05-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/repocheckoutservice/... ./hivepaas_app/service/syscleanupservice/... ./hivepaas_app/cmd/... && golangci-lint run ./...`
Expected: PASS; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(files): the repository cache is kept on the project's volume

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 6: A scheduled job's output

**Files:**
- Modify: `hivepaas_app/service/schedjobexecservice/schedjobexecserviceimpl/service.go`, `exec.go`, `exec_output.go`, `exec_output_to_file.go`
- Test: `hivepaas_app/service/schedjobexecservice/schedjobexecserviceimpl/exec_output_volume_test.go`

**Interfaces:**
- Consumes: Task 3's `Create`, `Remove`; Task 4's `ProjectVolume`, `AppFilePath`, `FilesDirJobOutput`.
- Produces: `schedjobexecserviceimpl.New(fileRepo, fileService, appService, commandService, containerExecService, schedJobService)`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-06-tests.patch`

Tests added: `TestAJobOutputIsWrittenToTheProjectVolume`, `TestAFailedJobLeavesNoOutputFile`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/service/schedjobexecservice/...`
Expected: FAIL, build error `unknown field fileService in struct literal of type service`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-06-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/service/schedjobexecservice/... ./hivepaas_app/cmd/... && golangci-lint run ./...`
Expected: PASS; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(files): a job's output file is kept on the project's volume

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 7: A volume holding files is not deleted

**Files:**
- Modify: `hivepaas_app/usecase/cluster/volumeuc/delete.go`, `hivepaas_app/hperrors/errors_infra.go` and its messages, `hivepaas_app/config/path.go`, `hivepaas_app/db/seed/seed.sql` (its three files become `volume`)
- Test: `hivepaas_app/usecase/cluster/volumeuc/delete_test.go`

**Interfaces:**
- Consumes: Task 3's `CountOnVolume`.
- Produces: `hperrors.ErrVolumeHasFiles` (`ERR_VOLUME_HAS_FILES`, with `Name` and `Count`).

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-07-tests.patch`

Tests added: `TestAVolumeHoldingFilesIsNotDeleted`, `TestAnEmptyVolumeIsDeleted`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/usecase/cluster/volumeuc/`
Expected: FAIL, build errors `undefined: refuseVolumeWithFiles` and `undefined: hperrors.ErrVolumeHasFiles`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-07-code.patch`

- [ ] **Step 4: Check everything**

Run: `go build ./... && go test ./... && golangci-lint run ./... && go run ./tools/goroutinelint . && go run ./tools/errcodelint`
Expected: PASS; `0 issues.` `make gen-swag` and `go generate ./hivepaas_app/interface/agent/proto/...` change nothing.

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(volume): a volume holding HivePaaS's files is not deleted

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 8: The dashboard, and the live checks

**Files** (under `$D/src/application/modules/projects/`):
- Modify: `domain/apps/data-file/app-data-file.entity.ts` (`Local` becomes `Volume: "volume"`), `api/services/project-apps-services/data-files/app-data-files.api.ts`, `module-shared/definitions/tables/app-data-files/app-data-files-table.defs.tsx` (reads "HivePaaS"), `…/building-blocks/app-data-file-menu-cell.com.tsx`

- [ ] **Step 1: Apply the patch**

Run: `git -C $D apply --index $(pwd)/$P/dash-01.patch`

- [ ] **Step 2: Check it**

Run: `(cd $D && npx tsc --noEmit && npm run lint && npx prettier --check src)`
Expected: no type errors, no lint output, `All matched files use Prettier code style!`

- [ ] **Step 3: Commit (dashboard repo)**

```bash
git -C $D commit -m "feat(files): an app's data file is on a HivePaaS volume

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

- [ ] **Step 4: Live checks** (the user, on the Linux server, with new app and agent images)
  - **Check 1:** on a database that is kept, `UPDATE files SET storage_type = 'volume' WHERE storage_type = 'local';` (Review Focus 2).
  - **Check 2:** deploy an app from a repository twice: the second deployment says the cache was found, and the archive is in `project_data/<project>/.hivepaas/cache/repos/`, mode `0700` on `.hivepaas`.
  - **Check 3:** a scheduled job saving its output to a file: the file is in `.hivepaas/job-output/<env>/<app>/` and downloads from the dashboard; note the env segment (Review Focus 5). Make the job fail: no file is left (Review Focus 4).
  - **Check 4:** upload a data file to an app and download it.
  - **Check 5:** with two nodes, pin a project's default volume to the other node and repeat checks 2 to 4 (Review Focus 1).
  - **Check 6:** delete a volume holding files: refused with the count. Make a project's default volume one that cannot hold files: the deployment warns and goes on (Review Focus 3).

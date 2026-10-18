# Managed Templates Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A user on a paid plan creates a Postgres 17 or 18 database from the
`postgres-managed` template. It archives WAL with WAL-G to S3, a cloud storage or a
volume repository, takes and prunes base backups, checks its chain, tests its own
restore, reports its health, and restores to a point in time into a new app. HivePaaS
builds no image for it: the official image runs with the WAL-G binary that the agent
downloaded, verified and mounted.

**Architecture:**
- **Schema (Tasks 1-2).** `v000002` adds `managed`, `artifacts` and
  `requires.features` to the template schema. The license payload gains `features`. A
  template that needs a feature the license lacks is shown, and cannot be created from.
- **Artifacts (Tasks 3-4).** A new agent RPC downloads a URL on its own node, hashes it
  while streaming, and keeps it in `<HostDir>/cache/artifacts/sha256/<hex>`. The backend
  ensures it on the app's pinned node before `ServiceCreate` and before every deploy,
  and adds a hidden read-only bind mount, treated as the Docker API socket mount is.
- **The engine (Tasks 5-9).**
  - `managedservice` holds one `Engine` per database, and `postgres` is the first.
  - Its state is the new `app-managed` setting.
  - It creates `container-command` jobs owned through `SchedJob.ManagedBy`, and refuses
    edits to the fields it owns.
  - Status is read from the jobs' last outputs.
- **Restore (Tasks 10, 11 and 11b).**
  - The restore test runs in the database's container or in a verifier app, a logical
    child scaled to zero between runs.
  - A restore creates a new app from the same template. It recovers from the source's
    archive, to the latest, a time, a restore point or a base backup, and archives to
    its own prefix.
  - Users name restore points, which are targeted by LSN. Retention keeps the kept ones
    restorable until their date.
- **Outside the backend (Tasks 12-14):**
  - the `postgres-managed` template and its scripts (`app-templates`);
  - the license portal issuing `features` (`hivepaas-portal`);
  - the dashboard (`hivepaas-dashboard`).

**Tech Stack:** Go 1.27, gRPC with protoc (as `repo_server.proto` was generated), WAL-G
3.0.8 (`wal-g-pg-20.04-{amd64,aarch64}`), Postgres 17 and 18 on Debian images, React
for the dashboard.

**Spec:** `docs/superpowers/specs/2026-10-17-managed-templates-design.md`.

## How this plan works

Unlike the plans before it, **no patches were written or verified in advance.** Each
task is written test-first when it is executed, and is done when its checks pass. The
file lists and interfaces below come from reading the code at this plan's commit. Where
the code has moved since, follow the code and note the change in the task's commit.

**Repositories:**
- this one (the backend), Tasks 1-11;
- `../app-templates`, Task 12;
- `../hivepaas-portal`, Task 13;
- `../hivepaas-dashboard`, Task 14.

Work on a branch per repository, in a worktree.

**Order:**
- Tasks 1 to 11 build on each other.
- Task 12 needs Task 1, plus Task 3 to run.
- Task 13 needs only Task 2.
- Task 14 needs the endpoints of Tasks 2, 7, 9, 10 and 11, and can start on their DTOs.

## Global Constraints

- **Before a task is done:** `go build ./...`, `golangci-lint run ./...` over the whole
  repo, `go test ./...`, and `make gen-swag` when a DTO or a handler's swag comment
  changed (CI fails on a stale `docs/openapi/swagger.json`). `go test
  ./hivepaas_app/cmd/...` includes the fx wiring test, which catches a dependency
  cycle.
- **Conventions:** `docs/ARCHITECTURE.md`. A service does not import a DTO. Request and
  response shapes follow §2. Switches over enums follow §7. Tests use `assert` and
  fakes.
- **Names, verbatim:**
  - Version code: `VersionCodeV2 = "v000002"`.
  - Setting types: `app-managed`.
  - Feature: `managed-databases`.
  - Engine: `postgres`, `engineApi` 1.
  - Errors: `ERR_PLAN_FEATURE_REQUIRED` (412) and `ERR_APP_MANAGED_FIELD_LOCKED` (412, under `ErrNonEditable` like the function rules),
    whose param is `Field`.
  - Host cache: `<HP_STORAGE_HOST_DIR>/cache/artifacts/sha256/<hex>`, files `0755`.
  - Container paths: artifacts under `/opt/hivepaas/bin/`.
  - WAL-G prefix: `walg/<app key>` under the chosen storage's prefix.
- **Limits:**
  - an artifact is at most 256 MiB;
  - downloads are https only (http only from a mirror on a private address, in dev);
  - `Ensure` times out after 10 minutes.
- **Secrets:** S3 keys and the libsodium key reach the container as Swarm secrets, never
  as env and never on argv. Logs name the step, not the value.
- **Git:** commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
  Merge locally, never push. Never `git stash`. Never touch the user's uncommitted work.
- **Deploys:** nothing is deployed to the user's Docker Desktop swarm. Live checks run
  on the user's Linux server.

## Review Focus

1. **A task starts before its artifact is there.**
   - Where this can happen: at creation, `ProvisionApp` calls `ServiceCreate` before
     the first deployment; and about forty other `ServiceUpdate` callers roll tasks.
   - Expectation: the artifact is on the pinned node before the service is created, and
     a managed app with an unpinned data volume is refused.
   - Pinned by: Task 4's create-order test, and Task 7's refusal test.
2. **Two servers writing one archive.** A restored app, and the restore test's copy,
   must never push to the source's prefix.
   - Expectation: the restored app archives to its own prefix, the test runs with
     `archive_mode=off`, and `WALG_PREVENT_WAL_OVERWRITE=true` is set everywhere.
   - Pinned by: Task 6's golden output, Task 10's argument test, Task 11's env test, and
     the live check in Task 11.
3. **Recovery that ends early and says nothing.** If `wal-fetch` fails with a network
   error and exits 1, Postgres treats the WAL as ended and promotes at an earlier point.
   - Expectation: `walg-fetch` turns every exit other than 74 into 255.
   - Pinned by: Task 12's script test, and the live check in Task 11.
4. **A license that lapses.**
   - Expectation: existing managed apps keep archiving and restore tests keep running;
     PITR and unmanage still work; only creating a managed app or adding a verifier is
     refused.
   - Pinned by: Task 2's and Task 11's tests.
5. **A user edits what the engine owns.**
   - Expectation: every such edit is refused with `ERR_APP_MANAGED_FIELD_LOCKED`. None is
     silently re-applied.
   - Pinned by: Task 8's per-use-case tests.

## Decisions beyond the spec

- **The restore test's verifier mode is its own job type,** `managed-restore-test`. It
  is a `var`, kept out of `AllSchedJobTypes` as `image-scan` is, because it has to
  scale the verifier up and down around the exec, and `container-command` needs a task
  that is already running. In container mode it is an ordinary `container-command`.
- **"Update tools" waits for phase 2.** Artifacts are fixed when the app is created. A
  new WAL-G reaches existing apps through phase 2.
- **Unmanage removes the jobs and keeps the archive.** The prefix and its data are
  left for the user to delete.

---

## Task 1: Templates can declare `managed`, `artifacts` and `requires.features`

**Files:**
- Modify:
  - `hivepaas_app/base/version.go`: `VersionCodeV2`, `CurrentVersion = VersionCodeV2`.
  - `hivepaas_app/service/apptemplateservice/templatemodel/template.go`:
    `Template.Managed *Managed`, `Template.Artifacts []*Artifact`,
    `Requires.Features []string`.
  - `.../templatemodel/validate.go`: `validateManaged`, `validateArtifacts`.
  - `.../templatemodel/index.go`: `IndexEntry.Managed string`, `IndexEntry.RequiresFeatures`.
  - `.../templaterepo/lint.go`: managed lint. `alpine` in a variant name or image tag
    is refused. `mountPath` must not be under a volume mount.
  - `tools/apptemplate/main.go`: `index` writes the new entry fields.
- Create:
  - `.../templatemodel/managed.go`: `Managed{Engine, EngineAPI, Postgres *ManagedPostgresDefaults}`,
    `Artifact{Name, Version, MountPath, Files map[base.CPUArch]*ArtifactFile}`,
    `ArtifactFile{URL, SHA256, Size}`.
  - `hivepaas_app/base/cpu_arch.go`: `CPUArch` (`amd64`, `arm64`) and
    `CPUArchFromDocker("x86_64"|"aarch64")`.
- Test: `templatemodel/managed_test.go` and lint fixtures under `templaterepo/testdata/`.

**Interfaces:**
- Produces: `templatemodel.Managed`, `templatemodel.Artifact`, `ArtifactFile`,
  `base.CPUArch` and `base.CPUArchFromDocker(string) (CPUArch, bool)`.
- **As built:** the engines a template may name are a table in the template format,
  `templatemodel.ManagedEngines` (`EngineContract{MinAPI, MaxAPI, Artifacts}`), and
  `Validate` checks against it. Task 5's registry is tested against the table, as the
  spec builders are against `BuildableBlocks`. This replaces the lint callback first
  planned.
- **As built:** artifacts may only go directly in `/opt/hivepaas/bin/`, so they cannot
  shadow a file of the image.

- [x] **Step 1:** Write the tests:
  - a valid managed template decodes and validates;
  - `KnownFields` still refuses an unknown key inside `managed` and `artifacts`;
  - each refusal: an unknown engine, `engineApi` out of range, an http URL, a hash that
    is not 64 hex characters, no `amd64`, a relative `mountPath`, a `mountPath` under a
    volume mount, an `alpine` tag;
  - lint refuses `v000002` blocks under `requires.versionCode: v000001`;
  - `subset -version-code v000001` leaves a managed template out.
- [x] **Step 2:** Run `go test ./hivepaas_app/service/apptemplateservice/... ./tools/apptemplate/...`. Expected: FAIL to build.
- [x] **Step 3:** Write the code.
- [x] **Step 4:** Run the task's packages and the global checks. Expected: all pass.
- [x] **Step 5:** Commit: `feat(templates): v000002 - managed, artifacts and requires.features`.

## Task 2: A license grants features, and a template that needs a missing one cannot be created from

**Files:**
- Modify:
  - `hivepaas_app/pkg/license/license.go`: `Payload.Features []string`.
  - `hivepaas_app/pkg/license/limits.go` (or beside it): Bee's compiled features, none.
  - `hivepaas_app/base/`: `LicenseFeature`, `LicenseFeatureManagedDatabases`.
  - `hivepaas_app/service/licenseservice/service.go`: `HasFeature(ctx, f) (bool, error)`.
  - `.../licenseserviceimpl/check.go`.
  - `hivepaas_app/hperrors`: `ErrPlanFeatureRequired`.
  - `hivepaas_app/usecase/apptemplateuc/app_create_from_template.go`: checks
    `entry.Requires.Features`, for the main template and every dependency.
  - `.../apptemplatedto/template_list.go` and `template_get.go`: `requiresFeatures` and
    `missingFeatures` in responses.
- Test: `pkg/license/evaluate_test.go`, `licenseserviceimpl/check_test.go` and
  `apptemplateuc/app_create_from_template_test.go`.

**As built:**
- `license.State.Features` and `Grants(feature)`, with `license.FeaturesOf` and
  `FeatureNameMsgID` in `pkg/license/features.go`.
- `licenseservice.CheckFeatures(ctx, features)` returns `ERR_PLAN_FEATURE_REQUIRED`,
  exposing `Feature` and `Plan`.
- `CreateAppFromTemplate` checks first, through `checkRequest`, which now holds every
  pre-provision check (gocognit).
- The list and get responses carry `managed`, `requiresFeatures` and
  `missingFeatures`.
- `tools/licensesign` gives every paid plan's key `managed-databases`, with
  `-features plan|none|a,b`, so that keys signed by hand match what Task 13 makes the
  portal issue.

- [x] **Step 1:** Write the tests:
  - a key without `features` grants none;
  - a Bee installation is refused `ERR_PLAN_FEATURE_REQUIRED`, with the feature as a
    param;
  - a key in grace still grants;
  - an expired key past grace grants none;
  - the list marks the missing features;
  - a key with an unknown feature name is still valid.
- [x] **Step 2:** Watch them fail.
- [x] **Step 3:** Write the code.
- [x] **Step 4:** Run the task's packages and the global checks.
- [x] **Step 5:** Commit: `feat(license): features, and templates that require them`.

## Task 3: The agent ensures an artifact on its node

**Files:**
- Create:
  - `hivepaas_app/interface/agent/proto/artifact.proto`:
    `ArtifactService.Ensure(EnsureReq{url, mirror_url, sha256, size}) returns (EnsureResp{host_path, downloaded bool})`.
    Generated `*.pb.go` files are committed.
  - `hivepaas_app/usecaseagent/artifactagentuc/ensure.go`: stat and size, then a hash
    once per agent start; otherwise download (mirror first) into a temp file beside the
    target, `io.TeeReader` into sha256, cap at `size` and 256 MiB, `chmod 0755`, then
    rename. A per-hash mutex makes concurrent calls download once.
  - `hivepaas_app/interface/agent/client/artifactservice/`.
  - `hivepaas_app/service/artifactservice/`: `EnsureOnNode(ctx, nodeID, *templatemodel.Artifact) (hostPath string, err error)`.
    It reads the node's architecture from Swarm with `base.CPUArchFromDocker`, picks the
    file, and calls the node's agent through `agentservice.GetAgentAddrForNode`.
- Modify:
  - `hivepaas_app/cmd/agent/main.go`: register the service.
  - `hivepaas_app/config/`: `HP_ARTIFACTS_MIRROR` (base URL, optional) and the timeout.
- Test: `artifactagentuc/ensure_test.go` against `httptest.Server`.

**As built:**
- The RPC is `ArtifactService.Ensure(EnsureArtifactReq) returns (EnsureArtifactResp)`.
- The agent's `artifactagentuc.Ensure` checks a file already there once per agent
  start. One that fails the check is removed and downloaded again.
- Only the source's error is returned. A failed mirror is logged.
- New errors: `ERR_ARTIFACT_INVALID`, `ERR_ARTIFACT_HASH_MISMATCH`,
  `ERR_ARTIFACT_DOWNLOAD_FAILED` and `ERR_ARTIFACT_ARCH_UNSUPPORTED`, which names the
  node and its architecture.
- `artifactservice.EnsureOnNode` reads the architecture from `NodeInspect` and asks
  the mirror for `<HP_ARTIFACTS_MIRROR>/<sha256>`.
- The protos are generated with protoc 35.1 (`v7.35.1` in their headers),
  `protoc-gen-go` v1.36.12 and `protoc-gen-go-grpc` v1.6.2, installed locally. The
  `hivepaas-devtools` image has protoc 31.1. Moving to a newer protoc waits until after
  the release.

- [x] **Step 1:** Write the tests:
  - a good file lands with mode 0755;
  - a wrong hash leaves no file and no temp file;
  - a connection cut halfway leaves nothing behind;
  - a file over `size` is refused while still downloading;
  - two concurrent `Ensure` calls cause one download;
  - a file already present is not downloaded again;
  - the mirror is tried first, and the source after the mirror fails;
  - an arch the template lacks is refused, naming the node and the arch.
- [x] **Step 2:** Watch them fail.
- [x] **Step 3:** Write the code, and generate the protos with `go generate ./hivepaas_app/interface/agent/proto/...`. The other protos are unchanged.
- [x] **Step 4:** Run the task's packages and the global checks.
- [x] **Step 5:** Commit: `feat(agent): ensure an artifact on a node by url and hash`.

## Task 4: An app can carry artifact mounts that HivePaaS owns

**Files:**
- Create: `hivepaas_app/service/artifactservice/mounts.go`, mirroring
  `dockerapiservice/attach.go`:
  - `ArtifactMounts(app) []mount.Mount`: bind, `ReadOnly`, **no `CreateMountpoint`**;
  - `KeepArtifactMounts(final, current)`;
  - `ApplyToService(ctx, db, app, spec)`.
- Modify:
  - `usecase/appsettingsuc/storage_settings_update.go`: keep them, as `KeepSocketMounts`.
  - `appsettingsdto/storage_settings_get.go`: hide them.
  - The storage settings version: leave them out.
  - `specserviceimpl/storage_map.go`: leave them out of export.
  - Image, repo and function deploy apply steps: `ensureArtifacts` beside
    `prepareDockerAPI`. Unlike it, **a failure fails the deploy**. `ApplyToService` runs
    inside `ServiceUpdateFunc`.
  - `appprovisionserviceimpl/provision.go`: the artifacts are ensured before
    `ServiceCreate`, through a `PreCreate` hook on the provision request.
  - `usecaseagent/nodecleanupagentuc`: remove cached artifacts that no app references
    after the cleanup's retention. The backend sends the referenced hashes.
- Test: one per touchpoint, plus `TestProvisionEnsuresArtifactsBeforeServiceCreate`.

**As built** (commits e267df37 and the cleanup after it). It differs from the list
above:
- **No `PreCreate` hook.** `ProvisionApp`'s `Configure` already runs before
  `ServiceCreate` and writes the spec. It runs inside the transaction, though, so Task
  7 ensures the artifacts *before* the transaction and `Configure` only attaches the
  mounts.
- **The artifacts live in their own setting, `app-artifacts`**
  (`entity.AppArtifactsSettings`), written from the template with
  `artifactservice.FromTemplate`. It is not part of `app-managed`, so the mechanism
  stays engine-agnostic. Task 5's `ManagedPostgres.Artifacts` is dropped. The setting
  is a singleton block `artifacts`, skipped by spec export and import.
- `artifactservice.MountsFor(artifacts, spec)` ensures them on every Ready, active
  node the spec's placement constraints allow. Nodes that would need different builds
  are refused, since one bind names one file. The deploy steps call it through
  `prepareArtifacts`, from the service's current placement.
- **Cleanup works like images do.** `clusterCleanupService.KeptArtifacts` reads the
  binds of the swarm's services, current and previous spec. `KeepArtifacts` and
  `ArtifactsKnown` go to each node (proto fields 9-10, output 16-17), and
  `artifactservice.PruneCache` removes files no service binds after the general
  retention. It runs only when images are pruned, and never on a request from an
  older manager.

- [x] **Step 1:** Write the tests:
  - storage GET hides the mount;
  - a storage save keeps it;
  - export leaves it out;
  - a deploy re-applies it;
  - a failed ensure fails the deploy, and no `ServiceUpdate` happens;
  - provision calls ensure before create;
  - cleanup never removes a referenced hash.
- [x] **Step 2:** Watch them fail.
- [x] **Step 3:** Write the code.
- [x] **Step 4:** Run the task's packages and the global checks.
- [x] **Step 5:** Commit: `feat(apps): artifact mounts that HivePaaS owns`.

## Task 5: The `app-managed` setting, owned jobs and the engine registry

**Files:**
- Create:
  - `hivepaas_app/entity/setting_app_managed.go`: `AppManagedSettings`, `ManagedPostgres`
    and `ManagedStorage{BackupRepo, CloudStorage, KeyAuth, Prefix}` (spec §3),
    `RestoreTest{Mode container|verifier, VerifierAppID, NodeID, CheckQuery}`
    (artifacts are in `app-artifacts`, Task 4), `CurrentAppManagedSettingsVersion = 1`, and `GetRefObjectIDs` (repo,
    storage, key and jobs).
  - `setting_app_managed_migration.go`, a no-op template.
  - `hivepaas_app/service/managedservice/service.go` and `managedmodel/engine.go`: the
    `Engine` interface (spec §3) and `Registry.Get(engine) (Engine, bool)`.
- Modify:
  - `base/setting.go`: `SettingTypeAppManaged`.
  - `entity/setting_sched_job.go`: `ManagedBy ObjectID`, with a ref link.
  - `usecase/settings/schedjobuc/update.go` and `delete.go`: refuse a job whose
    `ManagedBy` is set, with `ErrAppManagedFieldLocked{Field: "schedJob"}`.
  - The Task 1 lint stub becomes the registry.
- Test: entity round trip, the job refusals, and resource links that block deleting a
  storage a managed app uses.

**As built:**
- `entity.AppManagedSettings` holds `Engine` and `EngineAPI`, and
  `ManagedPostgres{Storage, EncryptionKey (encrypted), ArchiveTimeout, RetainFull,
  RestoreTest{Mode, VerifierApp, NodeID, CheckQuery}, Jobs{Nightly, Health,
  RestoreTest}}`.
- It references the repository, the storage, the key and the verifier app. It does
  not reference its jobs. Each job references it through `SchedJob.ManagedBy`, so
  removing the whole is not a cycle.
- `app-managed` is a singleton block `managed`, skipped by spec export and import.
- Update, status and delete of a managed job are refused through
  `refuseManagedJob`, with `ERR_APP_MANAGED_FIELD_LOCKED{Field: schedJob}`. Running
  one now is allowed.
- The job response carries `managedBy`.
- `managedmodel.Engine` starts with `Name` and `APIRange` only. Each task adds the
  methods it needs. The registry (`managedserviceimpl`) is tested against
  `templatemodel.ManagedEngines`.

- [x] **Step 1:** Write the tests. **Step 2:** Watch them fail. **Step 3:** Write the code. **Step 4:** Run the checks. **Step 5:** Commit: `feat(managed): the app-managed setting, owned jobs, the engine registry`.

## Task 6: The Postgres engine's `Configure` and `Apply`

**Files:**
- Create `hivepaas_app/service/managedservice/postgres/`:
  - `configure.go`:
    - the command arguments, from spec §3;
    - `shared_buffers` from the memory limit (25%, at least 128MB);
    - `PGDATA` per major (18: `/var/lib/postgresql/18/docker`);
    - the `WALG_*` env;
    - the storage resolved through `cloudstorageservice.S3Config`, or for a volume
      repository, a volume mount of `<prefix>/walg/<app key>` and `WALG_FILE_PREFIX`;
    - secrets for the keys and `WALG_LIBSODIUM_KEY` (generated, 32 bytes);
    - jobs: the nightly `job-sequence` (backup, retain, verify), health every 5 minutes,
      and the restore test, all with `ManagedBy` and failure notification by default;
    - the check that every script `engineApi` 1 names is among the template's config
      files.
  - `apply.go`: the arguments and artifact mounts again on every deploy.
  - `placement.go`: `off-site`, `shared` or `same-node`, for a volume repository, and
    the refusal of a volume pinned to another node.
- Test: `configure_test.go`, golden output per major (17, 18) per storage kind (S3
  repo, cloud storage, volume repo on the same node, volume repo unpinned).

**As built.** It differs from the list above:
- **Two steps instead of `Configure` and `Apply`:**
  - `PrepareDoc(PrepareReq)` adds to the rendered `AppDoc`, in blocks a template may
    use: the `WALG_*`, `AWS_*` and `HIVEPAAS_WALG_SECRETS_DIR` variables; the
    secrets `WALG_AWS_ACCESS_KEY_ID`, `WALG_AWS_SECRET_ACCESS_KEY` and
    `WALG_LIBSODIUM_KEY`, as files in `/run/secrets/walg`, uid 999, mode 400,
    referenced by no variable; and a volume repository's mount at
    `/var/lib/hivepaas-walg`. The spec builder then builds and checks it all, as it
    does a template's. It refuses a template missing one of the API's six scripts.
  - `Configure(ConfigureReq)` runs on the built settings. It sets the server's
    arguments on `app-deployment` (a template cannot), and returns `app-managed` and
    the three `container-command` jobs.
- **No `Apply`.** The arguments live in `app-deployment`, which every deploy
  applies, and Task 8 locks them.
- **The nightly job runs one script**, `walg-nightly`, which does backup, retain and
  verify. It is not a `job-sequence` of three jobs.
- **Resolving the archive moves to Task 7:** credentials, prefix, the volume
  repository's node and the placement warning. `PrepareDoc` takes it resolved, as
  `managedmodel.Archive`.

- [x] **Step 1:** Write the tests. The golden files show no secret value, and contain
  `archive_mode=on`, `WALG_PREVENT_WAL_OVERWRITE=true` and the prefix. Refused: a
  missing script, and a volume pinned to another node.
- [x] **Step 2:** Watch them fail. **Step 3:** Write the code. **Step 4:** Run the checks. **Step 5:** Commit: `feat(managed): the postgres engine - configure and apply`.

## Task 7: Creating a managed app from a template

**Files:**
- Modify:
  - `usecase/apptemplateuc/app_create_from_template.go`. When the rendered template has
    `managed`, it:
    - resolves the engine and its `engineApi`;
    - checks that the data volume is pinned (`ErrManagedVolumeNotPinned`);
    - checks the feature (Task 2);
    - passes `PreCreate`, which ensures the artifacts (Task 4);
    - merges the engine's `Configure` output into the `Configure` closure that
      `ProvisionApps` takes;
    - writes `app-managed`.
  - The create DTO gains `managed{storage{backupRepo|cloudStorage, keyAuth?},
    restoreTest{mode, nodeId?, checkQuery?}, archiveTimeout?, retainFull?}`, with
    defaults from `managed.postgres`.
- Test: `app_create_managed_test.go`.

**As built:**
- Before the transaction, `planManaged` resolves the plan:
  - the engine and its API;
  - the data volume's pin (`ERR_APP_MANAGED_VOLUME_NOT_PINNED`);
  - the archive. From an S3 repository the prefix is `<repo prefix>/walg/<app key>`;
    from a cloud storage it is `walg/<app key>`. An optional narrower `keyAuth` can be
    given. A volume repository counts as `same-node` or `shared`, and one pinned to
    another node is refused with `ERR_APP_MANAGED_ARCHIVE_UNREACHABLE`;
  - a 32-byte hex key, with `WALG_LIBSODIUM_KEY_TRANSFORM=hex` added to the engine's
    env;
  - the options, the request's over the template's defaults, with the cron
    schedules checked.

  It then runs `PrepareDoc` and downloads the artifacts with `MountsFor`, on a spec
  pinned like the data volume.
- Inside the transaction, `configureFromTemplate` runs the engine's `Configure`, adds
  the `app-artifacts` setting, and attaches the artifact mounts.
- The response carries `managed.placement` (`off-site`, `shared`, `same-node`). A
  managed template as a dependency is refused for now.
- The request gained `managed`, so **the API level went from 3 to 4**. The CLI has to
  be regenerated from the new spec and released before the server ships (see "After
  the plan"). MCP's install tools describe the new field.

- [x] **Step 1:** Write the tests:
  - success creates the app, the `app-managed` setting, three jobs and the artifact
    mount, in one transaction;
  - each refusal (feature, unpinned volume, engine unknown or out of range, storage of
    the wrong kind) creates nothing, and leaves no Swarm service;
  - a volume repository on the same node saves, with the warning in the response.
- [x] **Step 2:** Watch them fail. **Step 3:** Write the code. **Step 4:** Run the checks, and `make gen-swag`. **Step 5:** Commit: `feat(managed): create a managed app from its template`.

## Task 8: Fields the engine owns are refused, and an app can be unmanaged

**Files:**
- Create:
  - `usecase/appsettingsuc/managed_rules.go`, beside `function_rules.go`:
    `checkManagedKind`, `checkManagedDeployment` (the image repository is fixed; the
    tag may move only within its major, through `templatemodel`'s Postgres tag-family
    logic), `checkManagedContainer` (the engine's arguments come first and are
    unchanged), `checkManagedStorage` (the data and artifact mounts),
    `checkManagedEnvVars` (`WALG_*`, `AWS_*`, `PGDATA`).
  - `usecase/appmanageduc/unmanage.go`: remove `app-managed` and its jobs, keep the app
    and its data, and audit it.
- Modify: the kind, deployment, container, storage and env var settings update use
  cases call the rules, and the router gains `POST /apps/:appID/managed/unmanage`.
- Test: one refusal per rule, one allowed edit per rule (a minor tag bump, an extra
  argument, another env var), and unmanage.

**As built** (commit 204a4eab, and the unmanage commit after it):
- **The locks are data.** The engine records `entity.ManagedLocks{Image, ImageLine,
  Command, Mounts, EnvKeys, Settings}` in `app-managed` when it makes the app. The
  rules in `appsettingsuc/managed_rules.go` read them through `loadManagedLocks`
  and know nothing of Postgres.
- **Rules, all of which refuse:**
  - kind (category and engine);
  - deployment and container: method image; the repository kept; the tag within
    `ClassifyImageOverride`'s same line and the same `ImageTagFamily`; entrypoint
    and user unchanged; the engine's command a prefix, compared as argv;
  - storage: each locked target keeps type, source, subpath and writability;
  - env vars: each locked key keeps its entries exactly.
- **The engine's secrets, config files and setting mounts** are refused by
  update, status and delete in `usecase/settings` itself (`refuseManagedSetting`).
- **Unmanage** is `POST /apps/:appID/managed/unmanage`, in the new
  `appmanageduc` with `appmanagedhandler`. It soft-deletes `app-managed` and the
  jobs it made, cancels their runs, and is audited. Artifacts, secrets, scripts and
  arguments stay, so archiving goes on: removing them would leave
  `archive_command` failing and `pg_wal` filling the disk.

- [x] **Step 1:** Write the tests. **Step 2:** Watch them fail. **Step 3:** Write the code. **Step 4:** Run the checks, and `make gen-swag`. **Step 5:** Commit: `feat(managed): refuse edits to what the engine owns; unmanage`.

## Task 9: Status, and changing a managed app's settings

**Files:**
- Create `usecase/appmanageduc/`:
  - `status.go`: `GET /apps/:appID/managed/status`. It reads the last outputs of the
    health, nightly and restore-test jobs, as JSON printed by the scripts (Task 12):
    archiving `ok|failing|stale`, the size of `pg_wal`, the restore window, the last
    base backup, the last restore test, the archive's placement, and the suggestions.
    The verifier is suggested when free space is under 1.5 × the database's size, or
    the last test took over 30 minutes.
  - `update.go`: `PUT /apps/:appID/managed`. It changes schedules, retention, the
    restore test's mode and check query, and the key. The storage and prefix never
    change.
- Modify: `usecase/settings/` cloud storage and key auth updates mark the managed apps
  that link to them as needing a redeploy. Find them by resource links.
- Test: status from fixture outputs (none yet, healthy, failing, stale), each
  suggestion, an update of each field, and the redeploy mark.

**As built** (commits 596a78fc, 36775f06, c42b6ce2 and the credentials commit):
- **Outputs.** A managed job gets an output file of its run, `HIVEPAAS_OUTPUT`, as a
  sequence's step does. What it writes as KEY=value is kept as the task's output
  (`entity.TaskSchedJobOutputs`), whether the run failed or not.
- **`GET .../managed/status`.** It reads the last ended run of each of the engine's
  jobs; `postgres.Engine.Status` turns their outputs into the state. The outputs
  each script writes are named in the engine (`LAST_ARCHIVED_AT`, `FAILED_COUNT`,
  `PG_WAL_BYTES`, `DB_BYTES`, `FREE_BYTES`; `BACKUP_*`, `OLDEST_BACKUP_AT`;
  `RESTORED_TO`, `DURATION_SECONDS`). The placement is kept in
  `ManagedStorage.Placement`.
- **`GET` and `PUT .../managed`** cover retention, the two schedules and the check
  query, through `Engine.Reconfigure`. The scripts read their options from their
  job's env (`WALG_RETAIN_FULL`, `HIVEPAAS_RESTORE_CHECK_QUERY`). `archiveTimeout`
  is not changeable in phase 1: it is a locked server argument.
- **`POST .../managed/refresh-credentials`.** The credentials were copied into secrets
  when the app was made, so there is no mark on a key auth's update. Instead this
  re-resolves them through the new `managedservice.ResolveS3`, which the create flow
  now uses too. It rewrites only the secrets that changed (`Engine.ArchiveSecrets`)
  and mounts them again through `settingMountService.RecordRefresh`, which restarts
  the task.

- [x] **Step 1:** Write the tests. **Step 2:** Watch them fail. **Step 3:** Write the code. **Step 4:** Run the checks, and `make gen-swag`. **Step 5:** Commit: `feat(managed): status, and managed settings`.

## Task 10: The restore test in a verifier app

**Files:**
- Create:
  - `tasks/taskschedjobexec/managed_restore_test.go`: job type `managed-restore-test`
    (a `var`).
    - Container mode runs `walg-restore-test` through `schedJobExecService`.
    - Verifier mode scales the verifier to 1, waits for a running task (the existing
      `TaskFindRetry*`), runs the script in it, and always scales back to 0, in a
      `defer`.
  - `service/managedservice/postgres/verifier.go`: the verifier app is a logical child.
    It uses the same template version and artifact, its own volume on the node the
    user picks, scale 0, a read-only `KeyAuth` when one is given, and no archiving
    (`archive_mode=off` in its arguments).
- Modify:
  - Task 9's `update.go`: switching to verifier mode creates the verifier, and needs
    the feature; switching back deletes it.
  - `usecase/appuc/plan_limit.go`: a logical child is not counted, as the licensing
    design already says.
- Test: verifier creation (needs the feature); a run scales up and down, and also down
  when the script fails; the arguments contain `archive_mode=off`; a lapsed license
  still runs an existing verifier's test.

**As built:**
- **The move is a field of `PUT .../managed`:** `restoreTest{mode, volume, keyAuth}`.
  - To the verifier, it needs the feature, an archive in object storage
    (`ERR_APP_MANAGED_VERIFIER_NEEDS_OBJECT_STORAGE`), and a pinned volume, and
    refuses a second verifier (`ERR_APP_MANAGED_VERIFIER_EXISTS`).
  - The license check, the volume, the narrower key and the artifact download onto
    the volume's node all happen before the transaction (`planVerifier`).
  - Moving back asks no license, so a lapsed plan can always leave its verifier. It
    deletes the verifier last in the transaction, with its volume directory.
- **The verifier** (`provisionVerifier`) goes through `ProvisionApps`, which keeps
  its `Cleanup`:
  - name `<db> verifier`, a logical child of the database;
  - a copy of the database's env, secrets, config files, setting mounts, deployment
    and artifacts settings;
  - its secrets carry the narrower key's values through `Engine.ArchiveSecrets`;
  - its own directory of the volume at the engine's `ScratchDir`;
  - replicas 0, and a first deployment.
- **It runs no server.** Spec §5 step 1 needs the live database, and a verifier has
  none, so it idles (`Engine.Verifier()`: `sleep infinity`). There is no
  `archive_mode=off` in its arguments, because nothing in it archives. The server
  the script starts has archiving off (Task 12).
  - The job is told where it is: `HIVEPAAS_RESTORE_SOURCE_HOST` is the database's
    key on the project network, for the restore point, and
    `HIVEPAAS_RESTORE_SCRATCH_DIR` is where to restore. The script reads both
    (Task 12).
  - Pre- and post-deployment commands are cleared, because at 0 replicas there is no
    task to run them in.
- **The job** keeps its id and schedule. `Engine.Reconfigure` moves it, through
  `SettingsChange.RestoreTest`:
  - its type and kind become `managed-restore-test`, a `var`, so no request can
    ask for it;
  - its app becomes the verifier, and its env gets the two variables;
  - moving back undoes all three.
- **The executor** (`taskschedjobexec/managed_restore.go`):
  - It scales the verifier to 1 through a new `docker.Manager` dependency.
  - It runs the script there, waiting up to 24 × 5 s for a task.
  - It scales back to 0 in a `defer`, without the run's cancellation. A failure to
    scale down goes to the task's log.
  - It checks no license.
- **Kept with it:**
  - `ManagedRestoreTest.KeyAuth`. `refresh-credentials` refreshes the verifier's
    copies too, under that key (`archiveReaders`).
  - Unmanaging deletes the verifier.
  - Deleting the database deletes it too (`DeleteApp` cascades to logical
    children).
  - `plan_limit.go` needed no change: `CountApps` and `countedApps` already leave
    logical children out.
- **API level.** `gen-swag` raised it to 5. It was put back to 4: `PUT .../managed`
  is new in this branch, and level 3 is the released one.
- **Live check, deferred to Task 12** (it needs the scripts): on the Linux server,
  move a test database's restore test to a verifier on another node, run the job, and
  see it scale up, pass and scale down.

- [x] **Step 1:** Write the tests. **Step 2:** Watch them fail. **Step 3:** Write the code. **Step 4:** Run the checks. **Step 5:** Commit: `feat(managed): the restore test in a verifier app`.

## Task 11: Restore to a point in time into a new app

**Files:**
- Create:
  - `usecase/appmanageduc/restore.go`: `POST /apps/:appID/managed/restore{target{kind:
    latest|time|point|backup, time?, point?, backup?}, name?, dataVolume}`.
    - It checks that the target is inside Status's restore window, and that a named
      point or backup still exists.
    - It maps the target to its recovery settings (spec §4, "What to restore to"):
      - `latest`: none;
      - `time`: `recovery_target_time`;
      - `point`: `recovery_target_lsn` and `recovery_target_inclusive=on`;
      - `backup`: that backup with `recovery_target='immediate'`.

      For the other targets, `walg-entrypoint` fetches the latest base backup that
      starts before the target.
    - It creates the new app from the source's template binding (`app-template`
      setting: template, version and params) through Task 7's path, **without the
      feature check**.
    - The new app's env: `HIVEPAAS_RESTORE_SOURCE_PREFIX`,
      `HIVEPAAS_RESTORE_TARGET_TIME`, the source's read key, and its own prefix.
    - Its command goes through `walg-entrypoint`.
    - It inserts a `managed-restore` task.
  - `tasks/taskmanagedrestore/executor.go`. Steps, kept in the task with `Continue()`:
    1. wait for the service to run;
    2. wait for `pg_is_in_recovery() = false`, failing with Postgres's own message when
       recovery ends before the target;
    3. run `walg-backup` once;
    4. drop the restore env, and redeploy;
    5. mark the app ready.
- Test:
  - refusals: a time outside the window, a point or a backup that no longer exists, a
    target volume pinned to another node;
  - the recovery settings for each of the four targets;
  - the new app's env has its own prefix, and the source's only for fetching;
  - restore is allowed without the feature;
  - each step resumes after a crash;
  - the source app is not changed.

**As built:**
- **`POST .../managed/restore{target{kind, time?, point?, backup?}, name?, dataVolume}`.**
  - It checks write access on the source app and in its env, since it makes an app.
  - It checks the target against Status:
    - the window must exist (`ERR_APP_MANAGED_RESTORE_NO_BACKUP` before the first
      base backup);
    - a time must lie inside the window (`ERR_APP_MANAGED_RESTORE_OUTSIDE_WINDOW`,
      naming the window);
    - a point must be recorded and made after the window's start;
    - a backup must be one the last nightly run listed.
  - Status now lists the kept base backups, from a new nightly output, `BACKUPS`:
    `name@RFC3339` pairs, oldest first.
  - `ManagedPostgres.RestorePoints` exists now, so a point resolves to its LSN; Task
    11b records the points.
- **The new app** goes through `apptemplateuc.RestoreManagedApp`, which is Task 7's
  path with a `managedRestore`:
  - no feature check;
  - the engine is given the source (`PrepareReq.Restore`);
  - the `managed-restore` task is inserted in the same transaction and scheduled
    after the commit.
  - It is made from the source's binding:
    - its template, version and variant;
    - the tag of its image override;
    - its parameters, secrets decrypted, and the one volume parameter set to
      `dataVolume`. A template with more than one volume parameter is refused;
    - the same storage and key;
    - its options and schedules.
  - Its name defaults to `<app>-restored-<yyyymmdd-hhmm>`, and its prefix is its
    own, as any creation's is. It is not the source's child.
  - **The plan's app count still applies to a restore**, as the user decided on
    2026-10-18: a restore makes an app. Only the feature check is skipped.
- **The contract with the scripts.** It is engine API 1, and Task 12 implements it.
  - The source's archive is passed as the app's own variables, prefixed with
    `HIVEPAAS_RESTORE_` (`HIVEPAAS_RESTORE_WALG_S3_PREFIX`, `..._AWS_ENDPOINT`, or
    `..._WALG_FILE_PREFIX`).
  - The source's keys are files of `/run/secrets/walg/restore`
    (`HIVEPAAS_RESTORE_SECRETS_DIR`), and that includes the source's libsodium key.
  - A volume source is mounted read-only at `/var/lib/hivepaas-walg-source`.
  - `HIVEPAAS_RESTORE_TARGET_BACKUP` is the backup to fetch: a name, or `LATEST`.
    Without it, `HIVEPAAS_RESTORE_TARGET_BEFORE` gives a time, and the script fetches
    the latest backup that starts before it.
  - `HIVEPAAS_RESTORE_RECOVERY` holds the recovery settings as space-separated
    `k=v` pairs:
    - time: `recovery_target_time=<RFC3339>`;
    - point: `recovery_target_lsn=<lsn> recovery_target_inclusive=on`;
    - backup: `recovery_target=immediate`;
    - latest: none.
  - The engine validates LSNs and backup names before they reach a configuration
    file.
  - The app's entrypoint is `walg-entrypoint`, permanently. It passes through once
    `PGDATA` has data.
  - `walg-entrypoint status` exits 0 once the server is up and promoted, 75 while it
    is starting or recovering, and anything else on failure, after printing why. It
    prints nothing otherwise.
- **The jobs.** A restored app's engine jobs are created `disabled`, so health
  checks do not fail and notify during a recovery that takes hours. The managed
  loaders read `disabled` jobs too.
- **`tasks/taskmanagedrestore`.** Each run does one step, or waits on it for up to
  3 minutes, polling every 10 s, then `Continue()`s. The step is in the task's
  output.
  1. `recover` polls `walg-entrypoint status`. If the exec cannot start yet, it
     waits. It fails, non-retryable, on a failure exit or after 48 h.
  2. `backup` enables the engine's jobs and schedules them. It then inserts a run of
     the nightly job, scheduled after the commit, which Status reads as any other
     run.
  3. `wait-backup` waits for that run: done, or failed for good.
  4. `finish` calls `Engine.FinishRestore`:
     - it drops the `HIVEPAAS_RESTORE_*` variables, the `RESTORE_*` secrets with
       their mounts, the source mount, and their locks;
     - it sets `Restore.ReadyAt`;
     - then it applies the variables through `envvarservice`, removes the mount on
       the service, and records a setting-mount refresh.
  - The task has its own 15-minute timeout per run, and 5 retries a minute apart.
- **Status and settings.** `GET .../managed` shows `restore{sourceApp, target,
  readyAt}`. `GET .../managed/status` shows `backups`.
- **API level** stays 4: the restore is a new operation.
- **Live check, deferred to Task 12** (it needs the scripts).

- [x] **Step 1:** Write the tests. **Step 2:** Watch them fail. **Step 3:** Write the code. **Step 4:** Run the checks, and `make gen-swag`. **Step 5:** Commit: `feat(managed): restore into a new app - latest, a time, a point or a backup`.

## Task 11b: Restore points, and the retention that keeps them

**Files:**
- Create:
  - `usecase/appmanageduc/restore_point.go`:
    - `POST /apps/:appID/managed/restore-points{name, note?, keepUntil?}`. It checks the
      name: unique, at most 63 bytes, not `hp-`. It runs `pg_create_restore_point` and
      then `pg_switch_wal()` through `schedJobExecService.RunCommand`, waits for
      `pg_stat_archiver` to show the segment (timeout 2 × `archive_timeout`), records
      the point with the returned LSN, and is audited.
    - `PUT .../restore-points/:name{keepUntil|null, note}`.
    - `DELETE .../restore-points/:name`, which removes the record only.
  - `service/managedservice/postgres/retention.go`:
    `RetentionBoundary(backups []BaseBackup, retainFull int, points []RestorePoint, now)`
    returns the older of the n-th newest full backup and the full backup that starts
    before the oldest point still kept (`KeepUntil > now`). It returns none when there
    are fewer than n fulls.
  - The CLI's `hivepaas db mark <app> <name> [--keep-until] [--note]` lives in
    `../hivepaas-cli`, as a separate commit.
- Modify:
  - Task 6's nightly sequence: before `walg-retain` runs, the engine computes the
    boundary from `wal-g backup-list --json` (the previous step's output) and passes it
    as `walg-retain <backup>`.
  - Task 9's status: each point with the date until which it can be restored, and the
    storage held back by kept points.
  - `KeepUntil` is capped at 90 days by default.
- Test:
  - `RetentionBoundary` with no points, a point inside retention, a kept point older
    than retention, a lapsed point, and fewer fulls than n;
  - name refusals: a duplicate, too long, `hp-`;
  - a point is listed only after its segment is archived, and is refused on timeout;
  - deleting a point leaves the archive alone;
  - a lapsed license still allows creating points: they are part of restoring, which
    is never gated.

**As built:**
- **`POST .../managed/restore-points{name, note?, keepUntil?}`** runs the template's
  new script, `walg-mark <name> <seconds>`, in the database through `RunCommand`.
  - `RunCommandReq` gained a `StdoutWriter`, which is how the LSN comes back.
  - The script waits up to twice the archive timeout for the segment, then prints
    `LSN=<lsn>`. If the segment is not archived in time, it exits 75 and the point is
    refused with `ERR_APP_MANAGED_RESTORE_POINT_NOT_ARCHIVED`. Nothing is recorded.
  - Names: letters, digits, `.`, `_` and `-`, at most 63, not `hp-`, unique. The
    name goes into SQL, so the engine checks it too.
  - `keepUntil` is at most 90 days ahead, checked by the DTO.
  - No license is asked. Each change is audited.
- **`PUT .../restore-points/:name{keepUntil|null, note}`** sets both as given.
  **`DELETE .../restore-points/:name`** removes the record and runs nothing.
- **The nightly job is now type `managed-nightly`**, a `var`. The executor runs it
  in two steps of one run:
  1. `walg-nightly backup`, which prints `BACKUPS`;
  2. `Engine.Retention` decides the boundary, and the executor runs
     `walg-nightly retain [<boundary>]`.

  A failed backup is not followed by retention. The outputs are merged, with the
  retention's taking precedence.
  - This replaces the plan's "nightly sequence": the engine sits between the steps
    without a job sequence.
- **`RetentionBoundary`** is as planned, and counts only full backups (no `_D_`).
  The backup names the boundary is taken from are checked against the pattern of a
  WAL-G base backup, so an output cannot reach the command line.
- **Status** shows:
  - each point with `kept`, and `restorableUntil`: the nightly run, from its cron,
    that drops the backup before the point, while no kept point at or before it
    holds it;
  - `heldBack`, the full backups beyond `retainFull` and their bytes, from the
    `name@time@bytes` entries of `BACKUPS`. WAL is not counted.
- **The CLI.** `hivepaas db mark NAME [--keep-until] [--note]` uses the CLI's usual
  app selection (`-a` or the link), not a positional app.
  - It is in `../hivepaas-cli` on branch `feat/managed-db-mark` (worktree
    `../hivepaas-cli-managed`), commit caaefb2, based on `release`.
  - Its client is generated from this branch's spec (`SPEC_REF` is local, API
    level 4). Pin the release tag before releasing.
- **API level** stays 4: these are new operations.

- [x] **Step 1:** Write the tests. **Step 2:** Watch them fail. **Step 3:** Write the code. **Step 4:** Run the checks, and `make gen-swag`. **Step 5:** Commit: `feat(managed): restore points, and retention that keeps them`.

## Task 12 (`app-templates`): `postgres-managed`

**Files:**
- Create `templates/postgres-managed.yaml`:
  - `requires: {versionCode: v000002, features: [managed-databases]}`;
  - `managed: {engine: postgres, engineApi: 1, postgres: {...defaults}}`;
  - the `wal-g` artifact for amd64 and arm64. Take the sha256 and size from the
    release's `.sha256` files, then check them by downloading both;
  - versions 17 and 18 on pinned Debian images. 18 overrides the mount path, as
    `postgres.yaml` does;
  - the scripts as `settings.configFiles` at `/opt/hivepaas/bin/walg-*`, mode 0755:
    the engine's `apiScripts`, `walg-run`, `walg-fetch`, `walg-nightly`,
    `walg-health`, `walg-restore-test`, `walg-entrypoint` and `walg-mark`.
- Create `scripts/test/walg/`: shell tests run in `postgres:17` and `postgres:18`
  containers, with MinIO as the S3.
- Modify: `index.json` (via `apptemplate index`), and the README:
  - the new blocks and their version code;
  - fix the type count (9) and the dependency limit (5).

- **Contracts from Tasks 10 and 11, which the scripts must honor** (the engine's
  side is built):
  - `walg-nightly backup` takes the base backup, and prints `BACKUP_NAME`,
    `BACKUP_AT`, `BACKUP_BYTES` and `BACKUPS=<name@RFC3339@bytes ...>` (every
    backup kept, oldest first).
  - `walg-nightly retain [<backup>]` runs `delete before FIND_FULL <backup>
    --confirm` when given one, and nothing otherwise, then verifies the chain. It
    prints `OLDEST_BACKUP_AT` and `BACKUPS` again. It no longer reads
    `WALG_RETAIN_FULL` to decide anything.
  - `walg-mark <name> <seconds>`:
    - runs `pg_create_restore_point(name)`, then `pg_switch_wal()`;
    - waits until `pg_stat_archiver` shows the segment, at most `<seconds>`, else
      exits 75;
    - prints `LSN=<lsn>`.
  - `walg-restore-test` in a verifier:
    - takes its restore point on `HIVEPAAS_RESTORE_SOURCE_HOST`;
    - restores into `HIVEPAAS_RESTORE_SCRATCH_DIR`.
  - `walg-entrypoint` and `walg-fetch` take the restore's variables as Task 11's
    as-built notes say.
  - `walg-entrypoint status` exits 0, 75 or another code.
  - After a failed recovery, the entrypoint idles, so that `status` can report the
    failure, instead of the container restarting into another failure.
- [ ] **Step 1:** Write the script tests:
  - `walg-fetch` maps 1 to 255 and passes 0 and 74;
  - `walg-health` prints valid JSON, and fails on a grown `failed_count`;
  - `walg-restore-test` reaches its restore point, fails on a removed WAL segment, and
    refuses without free space;
  - `walg-entrypoint` recovers to each of the four targets: the latest, a time between
    two inserts, a restore point's LSN, and a base backup;
  - `walg-nightly retain <backup>` deletes only what is before that backup, and `walg-mark`
    waits for its segment, refusing on timeout.
- [ ] **Step 2:** Watch them fail. **Step 3:** Write the template and the scripts. **Step 4:** Run `apptemplate lint`, `index -check`, the compat job's subset for `v000001` (which leaves the template out), and the script tests. **Step 5:** Commit: `feat(templates): postgres-managed`.

## Task 13 (`hivepaas-portal`): keys carry features

- [ ] Every paid plan's key, Comb, Hive, Colony and Apiary, and the trial's, carries
  `features: ["managed-databases"]`. Bee has no key.
- [ ] A test of the issued payload per plan.
- [ ] Commit: `feat(keys): managed-databases on every paid plan`.

## Task 14 (`hivepaas-dashboard`): the managed screens

- [ ] **Store:**
  - a "Managed" badge;
  - the missing feature, with the plans that grant it, instead of the form.
- [ ] **Create form:** a Managed section with:
  - storage: an S3 repository, a cloud storage or a volume repository;
  - an optional separate key, with the reason for one;
  - the placement warning;
  - the restore test's mode and node;
  - schedules and retention.
- [ ] **The app's Database tab:**
  - Status: archiving, restore window, last backup and test, placement and
    suggestions;
  - Restore: the four targets (latest, a time bounded by the window, a restore
    point, a base backup) and the new volume;
  - Restore points: create with a note and an optional keep date, which shows the
    estimated storage; list each point with its "restorable until" date; delete;
  - Settings;
  - Unmanage, with a confirmation that names what goes.
- [ ] **Locked fields:** shown read-only with the reason, on every screen Task 8
  covers. The jobs list shows managed jobs as managed.
- [ ] e2e for create, status and restore against a mocked API. Lint and types are
  clean.
- [ ] Commit per screen.

## After the plan: release

- **The API level is 4** (Task 7). `hivepaas-cli` regenerates its client from the
  new spec and is released before the server ships, as `docs/DEVELOPMENT.md` asks.

## After the plan: live checks on the Linux server

1. Create managed PG17 and PG18 apps on a pinned volume, archiving to MinIO or an
   actual bucket. Write rows, and see `last_archived_time` move within 60 seconds.
2. Delete a row, then restore to a time before the delete into a new app. The row is
   there, and the source app is unchanged.
3. Mark a restore point, insert more rows, and restore to the point into a new app.
   It has exactly the rows from before the mark. Keep a point past `retainFull`
   nights, and see the base backup before it survive retention, and restoring to it
   still work.
4. The nightly sequence and the restore test, in both modes: they pass, and Status
   shows them.
5. Revoke the key. Archiver health fails within 5 minutes and notifies, and Status
   says `failing`.
6. Remove one WAL segment from the bucket. The next restore test fails on it.
7. Remove the license. Archiving and the restore test go on, and PITR still works
   while the apps are within Bee's count (a restore makes an app). Creating a managed
   app is refused.
8. On an arm64 node, if one is available: the aarch64 artifact is chosen and runs.

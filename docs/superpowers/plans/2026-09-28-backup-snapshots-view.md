# Backup Snapshots View Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A Backup Snapshots view at the global, project, env and app scopes lists the snapshots each scope may see, filtered by repository, app, tag and date, with details and delete.

**Architecture:**
- **Backend:** a new `backupsnapshotuc` use case.
  - **Reach.** It works out a view's reach - its repositories, the apps and repositories whose snapshots it holds, and what its viewer may read through `permission.Visibility`.
  - **List.** It lists the snapshot records of that reach with one SQL query over the `settings` and `tags` tables.
  - **Serving.** Handlers on the base settings handler serve it at the four scopes.
- **Engine:** kopia's "no snapshots matched" becomes `ErrSnapshotNotFound`, so a delete of a snapshot already gone counts as done.
- **Dashboard:** one `BackupSnapshotTable` serves the four pages.

**Tech Stack:** Go (bun, gin), PostgreSQL; React 19, TanStack Query and Table, zod.

**Spec:** `docs/superpowers/specs/2026-09-28-backup-snapshots-view-design.md`

## How the patches work

Every task's code was written and verified before this plan.

**Bases.** The backend patches apply in order to `main` at `974ba7c2`; this plan's commit changes only `docs/`. The dashboard patches apply to the dashboard's `main` at `6506753e`.

**What was checked:**

- **Backend, on a fresh worktree, at every task:**
  - the tests alone fail to build;
  - with the code, the task's packages and the fx wiring test pass, and `golangci-lint run ./...` finds 0 issues.
- **Backend, after the last task:** `go test ./...` passes.
- **The list query's SQL** was run, with `EXPLAIN` in a read-only session, against the local database: it is valid, and it uses the tags index.
- **Dashboard, after both patches:** `tsc`, `npm run lint` and `prettier --check src` pass.
- **The JSON the backend sends** was parsed with the dashboard's validator: a snapshot with an app, a job and a run; one with no tags; one whose app is deleted.

Patches live in `docs/superpowers/plans/2026-09-28-backup-snapshots-view/`:

- `be-NN-tests.patch`, then `be-NN-code.patch`, for backend task NN. Task 4 has code only: its gate is the fx wiring test and swagger.
- `dash-NN.patch` for dashboard task NN.

The commands below run from the backend repo root, with `P=docs/superpowers/plans/2026-09-28-backup-snapshots-view` and `D=../hivepaas-dashboard`.

## Global Constraints

- **Backend, before a task is done:**
  - `go build ./...`;
  - `golangci-lint run ./...` over the whole repo;
  - `go test ./...`.
  - `go test ./hivepaas_app/cmd/...` includes the fx wiring test.
- **`make gen-swag`** in Task 4, then `go mod verify`. A first run that fails with `cannot find all dependencies` is run again.
- **Dashboard:** `npx tsc --noEmit`, `npm run lint`, `npx prettier --check src`.
- **Main checkout:** after the merge, run `go test ./...` in the main checkout too. It has a local `vendor/` (git-ignored), so it builds with `-mod=vendor`. This plan imports nothing new.
- **Routes:**

  | Scope | Route |
  |---|---|
  | Global | `/settings/backup-snapshots` |
  | Project | `/projects/:projectID/backup-snapshots` |
  | Env | `/projects/:projectID/:projectEnv/backup-snapshots` |
  | App | `/projects/:projectID/:projectEnv/apps/:appID/backup-snapshots` |

  Each has `GET`, `GET /:itemID` and `DELETE /:itemID`.
- **Query parameters:**
  - `repo`, `app` and `tag` (several each, comma-joined as the dashboard's query builder sends them);
  - `fromDate` and `toDate` (`YYYY-MM-DD`, both included);
  - `search`, `pageOffset`, `pageLimit`.
  - A `sort` is ignored: the order is always newest first.
- **Tags:**
  - `hivepaas.app:<id>`, `hivepaas.job:<id>` and `hivepaas.run:<task id>`;
  - new, `hivepaas.source:command|volume`.
- **Git:**
  - commits end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`;
  - merge locally, never push;
  - never `git stash`, and never touch the user's uncommitted work.

## Review Focus

1. **A person with a grant on one env of a project** opens the project's view.
   - They must see only that env's apps' snapshots, and the project's repository's own snapshots - never another env's app, never another project's.
   - Pinned by: `TestProjectReach` (Task 2) for the reach, and Task 6, check 1, live with such a user.
2. **A global repository shared by two projects.**
   - Project A's view shows no snapshot of project B's apps.
   - Pinned by: `TestProjectReach` (Task 2), and Task 6, check 2.
3. **A large repository** (thousands of snapshots).
   - The view pages, filters by tag through the `tags` index, and does not load every snapshot.
   - Pinned by: the `EXPLAIN` check (done before this plan), and Task 6, check 3, on a repository with many snapshots.
4. **Deleting a snapshot someone already deleted with kopia.**
   - The delete succeeds, and the row goes.
   - Pinned by: `TestDeleteSnapshot_GoneIsNotFound` and `TestSnapshotGoneIsDeleted` (Task 3), and Task 6, check 4.
5. **A snapshot of a deleted app.**
   - It shows in its repository's scope as "Deleted app", and no longer in the app's, env's or project's view.
   - Pinned by: `TestTransformBackupSnapshotOfWhatIsGone` (Task 3) for the name, and Task 6, check 5.

## Decisions beyond the spec

These were taken while building the patches, and written back into the spec where they change it.

- **Dates:** the date filters are `fromDate` and `toDate`, as in the tasks list, and both are included.
- **The dashboard picks one repository and one app at a time,** while the API takes several. The app picker exists at the project and env views only, because there is no list of every app to pick from at the global view; a `hivepaas.app:<id>` tag narrows the global view to one app.
- **Delete** needs `delete` on the scope, as the other settings' deletes do, plus the owner rule of the spec's §1, and an active repository.
- **The project view reaches the project's envs' repositories too:** their own snapshots belong to the project's view when the viewer may read that env.
- **Records are listed without a scope filter,** by repository ID. A snapshot record copies its repository's scope but not its `inheritable` flag, so the scope filter would hide inherited repositories' snapshots.
- **The source kind** is read from the `hivepaas.source` tag, or from the job's data backup source for older snapshots.
- **Links in:**
  - the run page's snapshot ID links to the app's view filtered by `hivepaas.run:<task>`;
  - an app job list's data backup row gets a "Snapshots" link filtered by `hivepaas.job:<job>`.
- **The kopia engine** turns "no snapshots matched" into `backupmodel.ErrSnapshotNotFound`.

---

## Task 1: The source tag, and reading a snapshot's tags

**Files:**
- Modify: `hivepaas_app/entity/setting_sched_job_data_backup.go`, `hivepaas_app/service/databackupservice/databackupserviceimpl/backup.go`
- Test: `hivepaas_app/entity/setting_sched_job_data_backup_test.go`

**Interfaces:**
- Produces:
  - `entity.DataBackupTagSource = "hivepaas.source"` and `entity.DataBackupTagRun = "hivepaas.run"`;
  - `SnapshotTags` puts `hivepaas.source:<source>` after the app's tag;
  - `entity.DataBackupSnapshotTags{AppID, JobID, RunID, Source}`, from `entity.ParseDataBackupSnapshotTags(tags []string)`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-01-tests.patch`. `TestSchedJobDataBackupSnapshotTags` now expects the source tag; `TestParseDataBackupSnapshotTags` is added.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/entity/`
Expected: FAIL, build error `undefined: ParseDataBackupSnapshotTags`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-01-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test ./hivepaas_app/entity/ ./hivepaas_app/service/databackupservice/... && golangci-lint run ./...`
Expected: PASS; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(sched-jobs): a data backup's snapshots say their source, and their tags read back

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 2: A view's reach, and its query

**Files:**
- Create: `hivepaas_app/usecase/settings/backupsnapshotuc/reach.go`, `query.go`
- Test: `…/backupsnapshotuc/reach_test.go`, `query_test.go` (renders the SQL without a database)

**Interfaces:**
- Produces:
  - `snapshotReach{repoIDs, ownerRepoIDs, appIDs []string}`;
  - `computeReach(scope *entity.ObjectScope, repos []*entity.Setting, apps []*entity.App, allows func(projectID, env string) bool) *snapshotReach`;
  - `snapshotFilter{ID, RepoIDs, AppIDs, Tags, From, Before, Search}`;
  - `snapshotQueryOpts(reach, filter) []bunex.SelectQueryOption`.
- The query's shape:
  - `setting.type = 'backup-snapshot'`;
  - `setting.ref_id IN (reach ∩ filter repositories)`;
  - owned by an app of the reach (`EXISTS` on `tags`), or by no live app (`NOT EXISTS` joining `apps`) in a repository of the reach;
  - the filters;
  - `ORDER BY (setting.data->>'time')::timestamptz DESC`.
  - An empty reach, or a filter outside it, gives `1=0`.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-02-tests.patch`

Tests added:
- the reach: `TestGlobalReach`, `TestProjectReach`, `TestEnvReach`, `TestAppReach`;
- the query: `TestSnapshotQueryKeepsTheViewsReach`, `TestSnapshotQueryOfAnEmptyReach`, `TestSnapshotQueryFilters`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/usecase/settings/backupsnapshotuc/`
Expected: FAIL, build error `undefined: snapshotReach`

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-02-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go test -count=1 ./hivepaas_app/usecase/settings/backupsnapshotuc/ && golangci-lint run ./...`
Expected: PASS; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(backup): which snapshots a view reaches, and the query that lists them

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 3: List, get and delete

**Files:**
- Create:
  - `…/backupsnapshotuc/uc.go`, `fetch.go`, `list.go`, `get.go`, `delete.go`;
  - `…/backupsnapshotuc/backupsnapshotdto/list.go`, `get.go`, `transform.go`.
- Modify: `services/backup/kopia/delete.go` (not found)
- Test:
  - `…/backupsnapshotdto/transform_test.go`;
  - `…/backupsnapshotuc/delete_test.go`;
  - `services/backup/kopia/cmd_stderr_test.go`;
  - `hivepaas_app/interface/api/handler/basesettinghandler/backup_snapshot_request_test.go`.

**Interfaces:**
- Consumes: Task 1's tags, Task 2's reach and query.
- Produces, on `backupsnapshotuc.New(baseUC, appRepo, backupRepoService) *UC`:
  - `ListBackupSnapshot(ctx, auth, *ListBackupSnapshotReq) (*ListBackupSnapshotResp, error)`;
  - `GetBackupSnapshot(ctx, auth, *GetBackupSnapshotReq)`;
  - `DeleteBackupSnapshot(ctx, auth, *DeleteBackupSnapshotReq)`.
- Produces, in `backupsnapshotdto`:
  - `ListBackupSnapshotReq{Scope; RepoIDs "repo"; AppIDs "app"; Tags "tag"; FromDate, ToDate timeutil.Date; Search; Paging}`;
  - `ListBackupSnapshotResp{Meta, Data []*BackupSnapshotResp, Repos []*settings.BaseSettingResp}`;
  - `BackupSnapshotResp{ID, SnapshotID, ShortID, Time, SizeBytes, Description, Paths, Hostname, Tags, Source, Repo, App{ID, Name, Env, Deleted}, Job{ID, Name, Deleted}, RunID}`;
  - `TransformBackupSnapshot(record, tags, *SnapshotRefs)`.
- A snapshot outside the view's reach answers not found, as one that does not exist does.

- [ ] **Step 1: Write the failing tests**

Run: `git apply --index $P/be-03-tests.patch`

Tests added:
- the transform: `TestTransformBackupSnapshot`, `TestTransformBackupSnapshotOfWhatIsGone`;
- delete: `TestSnapshotGoneIsDeleted`, `TestDeleteSnapshot_GoneIsNotFound`;
- the request: `TestBackupSnapshotListRequestParses` (several repositories and tags, dates, search, page, a sort dropped), `TestBackupSnapshotListRequestRefusesABadTag`.

- [ ] **Step 2: Run them to watch them fail**

Run: `go test ./hivepaas_app/usecase/settings/backupsnapshotuc/... ./services/backup/kopia/ ./hivepaas_app/interface/api/handler/basesettinghandler/`
Expected: FAIL, build errors - the package `backupsnapshotdto` has no files yet.

- [ ] **Step 3: Write the code**

Run: `git apply --index $P/be-03-code.patch`

- [ ] **Step 4: Run them to watch them pass**

Run: `go build ./... && go test -count=1 ./hivepaas_app/usecase/settings/backupsnapshotuc/... ./services/backup/kopia/ ./hivepaas_app/interface/api/handler/basesettinghandler/ && golangci-lint run ./...`
Expected: PASS; `0 issues.`

- [ ] **Step 5: Commit**

```bash
git commit -m "feat(backup): list, read and delete the snapshots a view reaches

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 4: The four scopes' endpoints

**Files:**
- Create:
  - `hivepaas_app/interface/api/handler/basesettinghandler/xtra_api_backup_snapshot.go`;
  - `…/settinghandler/backup_snapshot.go`;
  - `…/projectsettingshandler/backup_snapshot.go`;
  - `…/projectenvsettingshandler/backup_snapshot.go`;
  - `…/appsettingshandler/backup_snapshot.go`.
- Modify:
  - `…/basesettinghandler/handler.go` (`BackupSnapshotUC`);
  - `hivepaas_app/interface/api/server/router_settings.go`, `router_projects.go`, `router_project_env.go`, `router_apps.go`;
  - `hivepaas_app/registry/provides.go`;
  - `docs/openapi/swagger.json`.

**Interfaces:**
- Consumes: Task 3's use case.
- The gate before the use case:
  - `GetAuthGlobalSettings` with `ResourceTypeBackupRepo` for the global scope, and the project, env and app settings checks for the others;
  - `read` to list and get, `delete` to delete.

- [ ] **Step 1: Write the code**

Run: `git apply --index $P/be-04-code.patch`. Then run `make gen-swag`, which must leave `swagger.json` unchanged, and `go mod verify`.

- [ ] **Step 2: Check it**

Run: `go build ./... && go test ./hivepaas_app/cmd/... ./hivepaas_app/interface/... && go test ./... && golangci-lint run ./...`
Expected: PASS, the fx wiring test included; `0 issues.` Also `grep -c backup-snapshots docs/openapi/swagger.json` gives `8`.

- [ ] **Step 3: Commit**

```bash
git commit -m "feat(backup): the backup snapshot endpoints at the global, project, env and app scopes

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 5: The Backup Snapshots views (dashboard)

**Files** (under `$D/src/application/`):
- Create:
  - `modules/settings/domain/backup-snapshot.entity.ts`;
  - `modules/settings/api/services/backup-snapshot-services/*`;
  - `modules/settings/api/hooks/use-backup-snapshot.api.ts`;
  - `modules/settings/data/queries/backup-snapshot.queries.ts`;
  - `modules/settings/data/commands/backup-snapshot.commands.ts`;
  - `modules/settings/module-shared/components/backup-snapshot-table/*` - the table, its columns, the details drawer and the delete dialog;
  - the pages: `modules/settings/routes/backup-snapshots/*`, `modules/projects/routes/single-project/configuration/backup-snapshots/*`, `modules/projects/routes/single-project/single-app/configuration/backup-snapshots/*`.
- Modify:
  - the indexes, the API context and the query keys;
  - `shared/constants/route.constants.ts`;
  - `settings.router.tsx`, `settings.module.ts`, `projects.router.tsx`, `projects.module.ts`, `routes/index.ts`;
  - the settings sidebar, the project sidebar, and the app's Configuration menu.

**Interfaces:**
- Consumes: Task 4's API.
- Produces:
  - `BackupSnapshotTable({ scope })`, where the scope is `{type: "settings"} | {type: "project", projectId, env?} | {type: "app", projectId, env, appId}`, reading `repo`, `app` and `tag` from the URL;
  - the routes `ROUTE.settings.backupSnapshots`, `ROUTE.projects.single.providerConfiguration.backupSnapshots` and `ROUTE.projects.single.apps.single.configuration.backupSnapshots`.

- [ ] **Step 1: Apply the patch**

Run: `git -C $D apply --index $(pwd)/$P/dash-01.patch`

- [ ] **Step 2: Check it**

Run: `(cd $D && npx tsc --noEmit && npm run lint && npx prettier --check src)`
Expected: no type errors, no lint output, `All matched files use Prettier code style!`

- [ ] **Step 3: Commit (dashboard repo)**

```bash
git -C $D commit -m "feat(backup): Backup Snapshots views at every scope

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

## Task 6: Links in, and the checks (dashboard, then live)

**Files:** `$D/src/application/modules/operations/routes/tasks/building-blocks/task-summary-card/system-task-summary-card.com.tsx`, and `$D/src/application/modules/projects/module-shared/definitions/tables/app-scheduled-jobs/building-blocks/view-tasks-cell.com.tsx`.

- [ ] **Step 1: Apply the patch**

Run: `git -C $D apply --index $(pwd)/$P/dash-02.patch`, then the checks of Task 5, Step 2.

- [ ] **Step 2: Commit (dashboard repo)**

```bash
git -C $D commit -m "feat(backup): a run's snapshot and a data backup's snapshots link to the app's view

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

- [ ] **Step 3: Live and browser checks** (the user, on the Linux server or the local stack):
  - **Check 1:** a user with a grant on one env of a project sees, in the project's view, that env's snapshots only (Review Focus 1).
  - **Check 2:** two projects backing up into one global repository - neither's view shows the other's (Review Focus 2).
  - **Check 3:** a repository with many snapshots pages and filters quickly (Review Focus 3).
  - **Check 4:**
    - delete a snapshot in an S3 repository and in a volume repository;
    - delete one already removed with `kopia snapshot delete` - it succeeds (Review Focus 4).
  - **Check 5:** a deleted app's snapshots show at the repository's scope as "Deleted app" (Review Focus 5).
  - **Check 6:**
    - the four pages, the filters and the details drawer;
    - the run page's snapshot link, and a data backup job's "Snapshots" link, open the app's view filtered.

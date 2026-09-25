# Setting Mounts, One Way - Inheritance Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Inheritable entries reach previews and clones.
- A preview resolves its parent's inheritable entries as its own, and is refreshed with its parent.
- A clone copies the source app's inheritable entries, pointing them at its own copies of the source app's settings.
- A clone leaves out entries with gated parts when the person who asked for it may not reveal them.

**Architecture:**
- **The engine.** `loadEntries` takes the app, and for a preview adds its parent's entries with `inheritable = TRUE`. Sources are read through the preview's own scope, which sees the parent's inheritable settings.
- **The refresh task.** After refreshing an app, it refreshes that app's previews.
- **The clone request.** `ExecuteAppClone` passes the Reveal Secrets gate for the source app's inheritable entries with gated parts. The answer travels to the clone task as `TaskAppCloneArgs.DropGatedMounts`.
- **The clone itself.** `cloneAppSettings` copies inheritable entries and remaps each one's source through a map from old to new setting id.

**Tech Stack:** Go, testify.

**Spec:** `docs/superpowers/specs/2026-09-25-setting-mounts-one-way-design.md` §4, and plan 2 of §5.

## Global Constraints

- **Entries:** `Setting.Inheritable` on an `app-setting-mount`. The API already takes it: `CreateSettingReq.Inheritable`, `UpdateSettingReq.Inheritable`.
- **Previews:**
  - a preview's entries are its parent's inheritable, active ones;
  - its Docker objects carry its own id and `GlobalKey`;
  - `RemoveApp` of a preview removes only its own.
- **Clones:**
  - an inheritable entry is copied;
  - its source is remapped to the copy when the source was one of the source app's settings and was copied;
  - it is dropped when that source was not copied;
  - a source outside the app keeps its id;
  - with `DropGatedMounts`, an entry with a gated part is dropped.
- **The gate at clone:**
  - asked on `uc.db`, recorded, once, only when an inheritable active entry has a gated part;
  - a denial is not an error: the clone goes on, and the response's `meta.warning` names the entries left out;
  - any other error fails the request.
- **Gates:** `go build ./...`, `golangci-lint run ./...`, `go test ./...`.
- **Git:**
  - branch `feat/setting-mounts-inheritance`;
  - every commit ends with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`;
  - merge locally, delete the branch, do not push.

## Review Focus

1. **A preview whose parent's entry mounts the parent's own secret sees that secret only when the secret is inheritable too.** Otherwise the entry's state is `source-unavailable`, not a crash. *Test: Task 1, `TestAPreviewResolvesItsParentsInheritableEntries`.*
2. **A clone of an app whose entry mounts a project secret keeps the id; one mounting its own secret gets the copy's id.** *Test: Task 2, `TestACloneCopiesInheritableEntriesWithTheirSources`.*
3. **An entry whose own secret was not cloned (`CloneSecrets` off) is dropped, not left pointing at the source app.** *Test: Task 2, same test.*
4. **A denied gate at clone does not fail the clone.** *Test: Task 3, `TestCloneRequestLeavesGatedMountsOutWhenDenied`.*
5. **Refreshing a parent refreshes its previews too, and a preview that fails does not stop the others.** *Test: Task 1, the executor test.*

---

### Task 1: Previews resolve their parent's inheritable entries

**Files:**
- Modify: `hivepaas_app/service/settingmountservice/settingmountserviceimpl/service.go`: `loadEntries` takes `*entity.App`, and `loadEntriesFromRepo` includes the parent's inheritable entries.
- Modify: `settingmountserviceimpl/resolve.go` (the call), `resolve_test.go` (the fixture's seam signature, and a new test).
- Modify: `hivepaas_app/tasks/tasksettingmountrefresh/executor.go` (refresh previews), `executor_test.go`.

**Interfaces:**
- Produces: `loadEntries func(ctx, db, app *entity.App) ([]*entity.Setting, error)`.

- [ ] **Step 1: Branch.** `git checkout main && git checkout -b feat/setting-mounts-inheritance`

- [ ] **Step 2: Write the failing tests.** In `resolve_test.go`, change the fixture's seam to take `*entity.App`. The fixture keeps returning the entries it was given; the parent's selection is the repository's (Step 4). Then add:

```go
// A preview's entries are its parent's inheritable ones; the files carry the
// preview's own name.
func TestAPreviewResolvesItsParentsInheritableEntries(t *testing.T) {
	inherited := entry(t, "cert", base.SettingStatusActive, certFiles("cert_1"))
	inherited.Inheritable = true
	svc := fixture(t, []*entity.Setting{inherited}, certSource(t, "cert_1", "CERT", "KEY"))
	var asked *entity.App
	loadAll := svc.loadEntries
	svc.loadEntries = func(ctx context.Context, db database.IDB, app *entity.App) ([]*entity.Setting, error) {
		asked = app
		return loadAll(ctx, db, app)
	}
	preview := &entity.App{ID: "app_preview", ParentID: testApp.ID, GlobalKey: "shop_prod_api-pr-7"}

	files, err := svc.Resolve(context.Background(), nil, preview)

	assert.NoError(t, err)
	assert.Equal(t, preview, asked, "the preview is who asks: its parent's entries are the repository's to add")
	assert.Len(t, files, 2)
}
```

In `executor_test.go`:
- give `fakeApps` a `List` that returns, for `parent_id` queries, the app `app_1_preview` when asked about `app_1`;
- the fake cannot read bunex options, so add a seam `listPreviews func(ctx, db, app) ([]*entity.App, error)` on `Executor`, and set it in the test;
- assert the refreshed order is `app_1, app_1_preview, app_broken, app_2`, and that `app_1_preview` counts as applied.

```go
func TestTheTaskRefreshesAnAppsPreviewsWithIt(t *testing.T) {
	mounts := &fakeMounts{}
	e := &Executor{appRepo: fakeApps{}, settingMountService: mounts, logger: logging.GlobalLogger()}
	e.listPreviews = func(_ context.Context, _ database.IDB, app *entity.App) ([]*entity.App, error) {
		if app.ID == "app_1" {
			return []*entity.App{{ID: "app_1_preview", ParentID: "app_1"}}, nil
		}
		return nil, nil
	}
	task := &entity.Task{Type: base.TaskTypeSettingMountRefresh}
	assert.NoError(t, task.SetArgs(&entity.TaskSettingMountRefreshArgs{AppIDs: []string{"app_1", "app_2"}}))

	assert.NoError(t, e.execute(context.Background(), database.Tx{}, &queue.TaskExecData{Task: task}))

	assert.Equal(t, []string{"app_1", "app_1_preview", "app_2"}, mounts.refreshed)
}
```

The existing executor test sets `listPreviews` to return none.

- [ ] **Step 3: Run** `go test ./hivepaas_app/service/settingmountservice/... ./hivepaas_app/tasks/tasksettingmountrefresh/`. It should FAIL to compile.

- [ ] **Step 4: Implement.** In `service.go`:

```go
	loadEntries func(ctx context.Context, db database.IDB, app *entity.App) ([]*entity.Setting, error)
```

```go
// loadEntriesFromRepo is the app's own entries, whatever their status, and for
// a preview its parent's inheritable ones: an entry is inherited only when
// whoever made it said so.
func (s *service) loadEntriesFromRepo(ctx context.Context, db database.IDB, app *entity.App) ([]*entity.Setting, error) {
	entries, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppSettingMount),
		bunex.SelectWhereGroup(
			bunex.SelectWhere("setting.object_id = ?", app.ID),
			bunex.SelectWhereOrIf(app.ParentID != "",
				"(setting.object_id = ? AND setting.inheritable = TRUE)", app.ParentID),
		),
	)
	return entries, hperrors.Wrap(err)
}
```

Check `bunex.SelectWhereOrIf`'s signature (`grep -n "func SelectWhereOrIf" hivepaas_app/pkg/bunex/*.go`); it is used by `setting_repo.go`. In `resolve.go`, call `s.loadEntries(ctx, db, app)`.

In `tasksettingmountrefresh/executor.go`:
- add the field `listPreviews func(ctx context.Context, db database.IDB, app *entity.App) ([]*entity.App, error)`, set in `NewExecutor` to `e.listPreviewsFromRepo`;
- after a successful `Refresh(app)`, refresh each preview the same way. Each counts in `Applied`, or in `Failed` under its own id:

```go
func (e *Executor) listPreviewsFromRepo(ctx context.Context, db database.IDB, app *entity.App) ([]*entity.App, error) {
	previews, _, err := e.appRepo.List(ctx, db, app.ProjectID, nil, bunex.SelectWhere("app.parent_id = ?", app.ID))
	return previews, hperrors.Wrap(err)
}
```

A preview has no previews of its own, so this does not recurse.

- [ ] **Step 5: Run** the two packages' tests, which should PASS. Then run `go build ./...` and `golangci-lint run ./hivepaas_app/...`.

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/service/settingmountservice/ hivepaas_app/tasks/tasksettingmountrefresh/
git commit -m "feat(settingmounts): a preview mounts its parent's inheritable entries, and is refreshed with it

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 2: Clones copy inheritable entries

**Files:**
- Modify: `hivepaas_app/service/appcloneservice/types.go` (`AppCloneReq.DropGatedMounts bool`)
- Modify: `hivepaas_app/service/appcloneservice/appcloneserviceimpl/`:
  - `clone.go`: `loadAppCloneData` reads `DropGatedMounts` from the task's args;
  - `clone_2_settings.go`: the default callback keeps inheritable entries, and `cloneAppSettings` remaps their sources after the loop.
- Modify: `hivepaas_app/entity/task_app_clone.go` (`DropGatedMounts bool \`json:"dropGatedMounts,omitempty"\``)
- Test: `clone_2_settings_test.go`

**Interfaces:**
- Produces:
  - `remapClonedMounts(entries []*entity.Setting, idMap map[string]string, srcOwn map[string]bool, dropGated bool) []*entity.Setting`, the pure part;
  - `TaskAppCloneArgs.DropGatedMounts`.

- [ ] **Step 1: Write the failing tests** in `clone_2_settings_test.go`. Replace `TestACloneLeavesSettingMountsBehind`, which pinned the old rule:

```go
func mountEntry(t *testing.T, name, source string, inheritable bool, parts ...string) *entity.Setting {
	t.Helper()
	mount := &entity.AppSettingMount{Source: entity.ObjectID{ID: source}}
	for _, part := range parts {
		mount.Files = append(mount.Files, &entity.AppSettingMountFile{Part: part, Path: "/etc/" + name + "/" + part})
	}
	setting := &entity.Setting{ID: "m-" + name, Type: base.SettingTypeAppSettingMount, Name: name,
		Status: base.SettingStatusActive, Inheritable: inheritable}
	assert.NoError(t, setting.SetData(mount))
	return setting
}

// A clone copies what its settings ask for, and an app's inheritable setting
// mounts: whoever made them said a copy may have them.
func TestTheDefaultCloneKeepsOnlyInheritableEntries(t *testing.T) {
	data := &appCloneData{AppCloneReq: &appcloneservice.AppCloneReq{CloneSettings: &entity.AppCloneSettings{}}}

	kept, err := (&service{}).onCloneSettingDefault(mountEntry(t, "conf", "cfg_1", true, "content"), data)
	assert.NoError(t, err)
	assert.NotNil(t, kept)
	dropped, err := (&service{}).onCloneSettingDefault(mountEntry(t, "key", "cert_1", false, "privateKey"), data)
	assert.NoError(t, err)
	assert.Nil(t, dropped)
}

// Each copied entry names the copy's setting where the source was the app's own
// and was copied; one whose own source was not copied is dropped; one mounting a
// setting from outside the app keeps it.
func TestACloneCopiesInheritableEntriesWithTheirSources(t *testing.T) {
	own := mountEntry(t, "conf", "cfg_1", true, "content")
	ownNotCopied := mountEntry(t, "token", "sec_1", true, "value")
	outside := mountEntry(t, "shared", "project_secret", true, "value")
	gated := mountEntry(t, "tls", "cert_1", true, "certificate", "privateKey")

	out := remapClonedMounts([]*entity.Setting{own, ownNotCopied, outside, gated},
		map[string]string{"cfg_1": "cfg_copy"},
		map[string]bool{"cfg_1": true, "sec_1": true},
		false)

	sources := map[string]string{}
	for _, setting := range out {
		sources[setting.Name] = setting.MustAsAppSettingMount().Source.ID
	}
	assert.Equal(t, map[string]string{"conf": "cfg_copy", "shared": "project_secret", "tls": "cert_1"}, sources)

	withoutGated := remapClonedMounts([]*entity.Setting{gated, outside}, nil, nil, true)
	assert.Len(t, withoutGated, 1, "a denied gate leaves the private key's entry out")
	assert.Equal(t, "shared", withoutGated[0].Name)
}
```

- [ ] **Step 2: Run** `go test ./hivepaas_app/service/appcloneservice/...`. It should FAIL.

- [ ] **Step 3: Implement.**
  - In `onCloneSettingDefault`, the `SettingTypeAppSettingMount` case returns `setting` when `setting.Inheritable`, else `nil`. Update the comment.
  - In `cloneAppSettings`:
    - build `idMap[setting.ID] = st.ID` for each copy, and `srcOwn[setting.ID] = true` for each of the source app's settings;
    - collect the copied entries apart;
    - after the loop, replace them in `data.ClonedSettings` with `remapClonedMounts(entries, idMap, srcOwn, data.DropGatedMounts)`.

  `remapClonedMounts`:

```go
// remapClonedMounts is the entries a clone keeps: each pointed at the copy of
// its source when the source was the app's own and was copied, dropped when it
// was the app's own and was not, kept as it is when it came from outside the
// app. With dropGated, an entry with a gated part is dropped: the person who
// asked for the clone may not reveal it.
func remapClonedMounts(
	entries []*entity.Setting, idMap map[string]string, srcOwn map[string]bool, dropGated bool,
) []*entity.Setting {
	kept := make([]*entity.Setting, 0, len(entries))
	for _, setting := range entries {
		mount, err := setting.AsAppSettingMount()
		if err != nil {
			continue
		}
		if dropGated && len(settingmountservice.Grants(mount)) > 0 {
			continue
		}
		if srcOwn[mount.Source.ID] {
			copied, ok := idMap[mount.Source.ID]
			if !ok {
				continue
			}
			mount.Source.ID = copied
			if err = setting.SetData(mount); err != nil {
				continue
			}
		}
		kept = append(kept, setting)
	}
	return kept
}
```

  - In `loadAppCloneData`'s task branch, set `data.DropGatedMounts = taskArgs.DropGatedMounts`.
  - Add the task argument, and `AppCloneReq.DropGatedMounts` with a comment: "leaves out setting mounts with a gated part: the clone's requester may not reveal them".

- [ ] **Step 4: Run** `go test ./hivepaas_app/service/appcloneservice/... ./hivepaas_app/tasks/...`. It should PASS. Then run `golangci-lint run ./hivepaas_app/...`.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app/service/appcloneservice/ hivepaas_app/entity/task_app_clone.go
git commit -m "feat(settingmounts): a clone copies inheritable entries, pointed at its own copies of their sources

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 3: The gate at the clone request

**Files:**
- Modify: `hivepaas_app/service/appcloneservice/service.go` and `clone_task_create.go`: `CreateAppCloneTask(app *entity.App, dropGatedMounts bool)`.
- Modify: `hivepaas_app/usecase/appsettingsuc/clone_execute.go`:
  - a `cloneMountsGate(ctx, auth, app) (drop bool, leftOut []string, err error)` before the transaction;
  - the warning in `ExecuteAppCloneResp.Meta`.
- Test: `hivepaas_app/usecase/appsettingsuc/clone_execute_test.go` (create)

**Interfaces:**
- Consumes: `settingmountservice.Grants`, `permission.Manager.AuthorizeSecretReveal`, `TaskAppCloneArgs.DropGatedMounts`.

- [ ] **Step 1: Write the failing test.** `clone_execute_test.go` tests the gate helper with fakes for the setting repository's `List` and the permission manager. Read `container_settings_get_test.go` in the same package for how its fakes are built, and follow it:

```go
func TestCloneRequestLeavesGatedMountsOutWhenDenied(t *testing.T) {
	for name, tc := range map[string]struct {
		err      error
		wantDrop bool
	}{
		"allowed":               {nil, false},
		"no capability":         {hperrors.ErrUserNotHavePermissionOnRevealSecrets, true},
		"secrets not returned":  {hperrors.ErrRevealSecretsDisabled, true},
	} {
		perms := &fakeRevealPerms{err: tc.err}
		uc := &UC{permissionManager: perms}
		entries := []*entity.Setting{
			inheritableEntry(t, "tls", "cert_1", "certificate", "privateKey"),
			inheritableEntry(t, "conf", "cfg_1", "content"),
		}

		drop, leftOut, err := uc.cloneMountsGate(context.Background(), nil, &entity.App{ID: "app_1"}, entries)

		assert.NoError(t, err, name)
		assert.Equal(t, tc.wantDrop, drop, name)
		if tc.wantDrop {
			assert.Equal(t, []string{"tls"}, leftOut, name)
		}
		assert.Len(t, perms.subjects, 1, "asked once, recorded")
	}
}

func TestCloneRequestAsksNothingWithoutGatedParts(t *testing.T) {
	perms := &fakeRevealPerms{}
	uc := &UC{permissionManager: perms}

	drop, _, err := uc.cloneMountsGate(context.Background(), nil, &entity.App{ID: "app_1"},
		[]*entity.Setting{inheritableEntry(t, "conf", "cfg_1", "content")})

	assert.NoError(t, err)
	assert.False(t, drop)
	assert.Empty(t, perms.subjects)
}
```

The helper takes the entries already loaded, so the test needs no repository. `ExecuteAppClone` loads them: `uc.settingRepo.List(ctx, uc.db, nil, nil, type = app-setting-mount, object_id = app, inheritable = TRUE, status = active)`.

- [ ] **Step 2: Run** `go test ./hivepaas_app/usecase/appsettingsuc/`. It should FAIL.

- [ ] **Step 3: Implement.**

```go
// cloneMountsGate passes §7's gate for the gated parts of the app's inheritable
// entries, which a clone would copy: the person asking for the clone gets an app
// that reads them. A denial is an answer, not an error - the clone goes on
// without those entries, and says so.
func (uc *UC) cloneMountsGate(
	ctx context.Context, auth *basedto.Auth, app *entity.App, entries []*entity.Setting,
) (drop bool, leftOut []string, err error) {
	var grants []settingmountservice.Grant
	for _, setting := range entries {
		mount, parseErr := setting.AsAppSettingMount()
		if parseErr != nil {
			return false, nil, hperrors.Wrap(parseErr)
		}
		if g := settingmountservice.Grants(mount); len(g) > 0 {
			grants = append(grants, g...)
			leftOut = append(leftOut, setting.Name)
		}
	}
	if len(grants) == 0 {
		return false, nil, nil
	}
	detail, err := json.Marshal(map[string]any{"grants": grants})
	if err != nil {
		return false, nil, hperrors.Wrap(err)
	}
	err = uc.permissionManager.AuthorizeSecretReveal(ctx, uc.db, auth, &permission.RevealSubject{
		Scope: base.ObjectScopeApp, ObjectID: app.ID, Source: base.AuditLogSourceAPIAction,
		ResType: base.ResourceTypeSettingMount, ResID: app.ID, ResName: "clone (setting mounts)",
		Detail: string(detail),
	})
	switch {
	case err == nil:
		return false, nil, nil
	case errors.Is(err, hperrors.ErrRevealSecretsDisabled),
		errors.Is(err, hperrors.ErrUserNotHavePermissionOnRevealSecrets):
		return true, leftOut, nil
	}
	return false, nil, hperrors.Wrap(err)
}
```

In `ExecuteAppClone`, before the transaction:
- load the app's inheritable active entries;
- call the gate;
- pass `drop` to `CreateAppCloneTask`, threading it through `executeAppCloneData` into `loadAppCloneSettingsForExecute`;
- when `leftOut` is non-empty, set `resp.Meta = &basedto.Meta{Warning: "The clone was made without these setting mounts, whose files you may not reveal: " + strings.Join(leftOut, ", ")}`.

`CreateAppCloneTask` sets `DropGatedMounts` in the args. `uc.db` is `*database.DB` and satisfies `database.IDB`; `nil` in the test stands for it.

- [ ] **Step 4: Run** `go build ./... && go test ./hivepaas_app/usecase/appsettingsuc/... ./hivepaas_app/service/appcloneservice/... ./hivepaas_app/cmd/...`. It should PASS. Then run `golangci-lint run ./...`.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app/usecase/appsettingsuc/ hivepaas_app/service/appcloneservice/
git commit -m "feat(settingmounts): a clone request passes the reveal gate for the entries it would copy

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

### Task 4: Gates, spec, merge

- [ ] **Step 1: Gates.** `go build ./...`, `golangci-lint run ./...`, `go test ./...`.
- [ ] **Step 2: Update the spec.** In the one-way spec's §5, plan 2 is built. Note that `inheritable` needed no API change: `CreateSettingReq`/`UpdateSettingReq` carry it.
- [ ] **Step 3: Commit, merge, delete the branch.**

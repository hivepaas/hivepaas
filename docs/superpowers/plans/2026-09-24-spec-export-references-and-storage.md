# Spec Export: References and Storage Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Export writes every reference and every mount in a form import can resolve on any
installation: a reference that leaves the export as an external reference, and a mount into
an app's directory the way the storage screen and a template write it.

**Architecture:** Two export-side changes, no reader yet. (1) Before documents are written, the
settings that exported settings reference but the export does not hold are loaded and
registered as external references; the existing `replaceExternalRefs` pass then writes them.
(2) An app's storage is mapped from its Swarm mounts with the help of
`volumeservice.DescribeAppMounts`, which learns to report the volume each mount reaches: a
mount into an app's directory becomes a managed mount (volume by path or external reference,
subpath below the app's directory, owning app when it is not this one), every other mount a
Docker mount kept as Docker holds it.

**Tech Stack:** Go 1.27, moby API types, bun, gopkg.in/yaml.v3, testify.

**Spec:** [docs/superpowers/specs/2026-09-23-config-spec-import-design.md](../specs/2026-09-23-config-spec-import-design.md)
- §9 "Storage is the one block export writes in a form the builder cannot read"; §4.

## Global Constraints

- Before calling a task done: `go build ./...`, `golangci-lint run ./...` over the whole repo
  (120-character lines, US spelling), and the tests the task names. The last task runs
  `go test ./...`.
- `storage.mounts` keeps exactly the form a template writes - `type`, `source`, `readOnly`,
  `volumeOptions: {subpath, noCopy}`, `sourceApp` - so `CheckBuildable` and the template
  builder are unchanged. Export adds `consistency` and `clusterOptions` to it only when Docker
  holds them.
- A managed mount's `source` is the volume's scope path when the export holds the volume. When
  it does not, `source` is empty and `external: {type, name, kind, id}` names it.
- `storage.dockerMounts` holds every other mount exactly as `mapMount` writes one today.
- The shared-memory mount (`tmpfs` at `/dev/shm`) appears in neither map:
  `resources.memory.shmSize` carries it.
- A target appears at most once across both maps.
- Another app's directory is named only inside the mounting app's environment.
- A reference to a setting outside the export is written as `{external: {type, name, kind, id}}`
  in place of the id string, by the existing `replaceExternalRefs`.
- Two exports of unchanged data stay byte-identical apart from `exportedAt`.
- End every commit message with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

---

## File Structure

| file | change |
|---|---|
| `hivepaas_app/service/volumeservice/types.go` | `AppMountDesc.VolumeID` |
| `hivepaas_app/service/volumeservice/volumeserviceimpl/app_mount_describe.go` | report the volume; name another app only in the same environment |
| `hivepaas_app/service/volumeservice/volumeserviceimpl/app_mount_describe_test.go` | tests for both |
| `hivepaas_app/service/specservice/specmodel/deployment.go` | `Storage.DockerMounts`, `Mount.External` |
| `hivepaas_app/service/specservice/specmodel/buildable_test.go` | a template may use neither |
| `hivepaas_app/service/specservice/specserviceimpl/storage_map.go` | new: `mapAppStorage` |
| `hivepaas_app/service/specservice/specserviceimpl/storage_map_test.go` | new |
| `hivepaas_app/service/specservice/specserviceimpl/swarm_map.go` | storage leaves `mapSwarmService` |
| `hivepaas_app/service/specservice/specserviceimpl/swarm_map_test.go` | storage tests move out; the test service gains two mounts |
| `hivepaas_app/service/specservice/specserviceimpl/service.go` | `loadByIDs` seam |
| `hivepaas_app/service/specservice/specserviceimpl/walk.go` | external references; storage through `DescribeAppMounts`; `db` threaded to the app documents |
| `hivepaas_app/service/specservice/specserviceimpl/export_test.go` | fixture: volumes, a volume service, project-scope exports |
| `hivepaas_app/service/specservice/specserviceimpl/export_refs_test.go` | new: the export tests of this plan |
| `docs/superpowers/specs/2026-09-23-config-spec-import-design.md` | §9, corrections, order of work |

All commands run from `/Users/tnt/go/src/github.com/hivepaas/hivepaas`.

---

### Task 1: DescribeAppMounts reports the volume, and names another app only in its environment

**Files:**
- Modify: `hivepaas_app/service/volumeservice/types.go` (`AppMountDesc`)
- Modify: `hivepaas_app/service/volumeservice/volumeserviceimpl/app_mount_describe.go`
- Test: `hivepaas_app/service/volumeservice/volumeserviceimpl/app_mount_describe_test.go`

**Interfaces:**
- Consumes: `describeAppMount`, `appStorageTarget`, `appScopePrefix`, `appKeyInSubpath` (all
  existing, in `volumeserviceimpl`); test helpers `mountTestApp()` (project `shop`, env `prod`,
  app `web`), `scopedVolume(t, id, refID, scope, vol)`, `clusterVolumeSetting(t, name, device)`
  (project scope, id = name).
- Produces: `volumeservice.AppMountDesc.VolumeID string` - the cluster-volume setting's id,
  set whenever `AppKey` is.

- [ ] **Step 1: Write the failing tests**

In `app_mount_describe_test.go`, `TestDescribeAppMountNamesTheOwner` gains one assertion after
`assert.Equal(t, tc.wantSub, desc.Subpath)`:

```go
			assert.Equal(t, "vol-1", desc.VolumeID)
```

`TestDescribeAppMountReadsABind` gains, after `assert.False(t, desc.Own)`:

```go
	assert.Equal(t, "vol-1", desc.VolumeID)
```

`TestDescribeAppMountLeavesTheUnknownUnnamed` gains, after `assert.False(t, desc.Own)`:

```go
			assert.Empty(t, desc.VolumeID)
```

and append:

```go
// An app is found again by its key only inside its own environment - the same
// key in another environment is another app. A directory in a volume shared
// wider than one environment is named only when it lies in this app's.
func TestDescribeAppMountNamesNoAppOutsideItsEnvironment(t *testing.T) {
	app := mountTestApp() // project shop, env prod, app web
	projectVolume := scopedVolume(t, "vol-p", "hp-vol-p", base.ObjectScopeProject, &entity.ClusterVolume{Managed: true})
	globalVolume := scopedVolume(t, "vol-g", "hp-vol-g", base.ObjectScopeGlobal, &entity.ClusterVolume{Managed: true})

	cases := map[string]struct {
		source  string
		subpath string
		wantKey string
	}{
		"a project volume, this environment":    {"hp-vol-p", "prod/postgres", "postgres"},
		"a project volume, another environment": {"hp-vol-p", "staging/postgres", ""},
		"a global volume, this environment":     {"hp-vol-g", "shop/prod/postgres", "postgres"},
		"a global volume, another environment":  {"hp-vol-g", "shop/staging/postgres", ""},
		"a global volume, another project":      {"hp-vol-g", "blog/prod/postgres", ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			desc := describeAppMount(app, &mount.Mount{
				Type: mount.TypeVolume, Source: tc.source, Target: "/data",
				VolumeOptions: &mount.VolumeOptions{Subpath: tc.subpath},
			}, []*entity.Setting{projectVolume, globalVolume})

			assert.Equal(t, tc.wantKey, desc.AppKey)
			if tc.wantKey == "" {
				assert.Empty(t, desc.VolumeID, "a directory nobody is named for names no volume either")
			}
		})
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./hivepaas_app/service/volumeservice/volumeserviceimpl/ -run TestDescribeAppMount -v`

Expected: the package does not compile - `desc.VolumeID undefined (type
*volumeservice.AppMountDesc has no field or method VolumeID)`.

- [ ] **Step 3: Add the field**

In `hivepaas_app/service/volumeservice/types.go`, `AppMountDesc` gains, after `Subpath`:

```go
	// VolumeID is the cluster-volume setting whose directory the mount reaches.
	// It is set whenever AppKey is: a mount that is no app's directory names no
	// volume here either.
	VolumeID string
```

- [ ] **Step 4: Report the volume, and keep another app inside the environment**

In `hivepaas_app/service/volumeservice/volumeserviceimpl/app_mount_describe.go`,
`describeAppMount` becomes:

```go
func describeAppMount(
	app *entity.App,
	mnt *mount.Mount,
	volumes []*entity.Setting,
) *volumeservice.AppMountDesc {
	desc := &volumeservice.AppMountDesc{}

	target, ok := appStorageTarget(mnt, volumes)
	if !ok || target.volume == nil {
		return desc // a volume mounted whole, or one nothing here accounts for
	}
	scope := target.volume.Scope

	if appOwnsSubpath(app, scope, target.subpath) {
		desc.AppKey, desc.Own = app.Key, true
		desc.Subpath = trimDirPrefix(target.subpath, appScopePrefix(app, scope))
		desc.VolumeID = target.volume.ID
		return desc
	}

	key, rest, ok := appKeyInSubpath(scope, target.subpath)
	if !ok || !inAppsEnvironment(app, scope, target.subpath) {
		return desc
	}
	desc.AppKey, desc.Subpath, desc.VolumeID = key, rest, target.volume.ID
	return desc
}
```

and below `appKeyInSubpath`:

```go
// inAppsEnvironment reports whether a directory inside a volume of this scope
// lies in the app's own environment - below <project>/<env> in a global volume,
// below <env> in a project's. An app is found again by its key only there: the
// same key in another environment is another app, and naming it would hand one
// app's files to a stranger. An environment's volume and an app's hold that
// environment alone.
func inAppsEnvironment(app *entity.App, scope base.ObjectScopeType, subpath string) bool {
	prefix := appScopePrefix(app, scope)
	if prefix == "" {
		return false
	}
	envDir := filepath.Dir(prefix)
	if envDir == "." {
		return true
	}
	subpath = strings.TrimPrefix(filepath.Clean(subpath), "/")
	return strings.HasPrefix(subpath, envDir+"/")
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./hivepaas_app/service/volumeservice/... ./hivepaas_app/usecase/appsettingsuc/... -v -run 'TestDescribeAppMount|Storage'`

Expected: PASS. The storage screen reads these descriptions too, and its tests must still pass.

- [ ] **Step 6: Lint and commit**

```bash
go build ./... && golangci-lint run ./...
git add hivepaas_app/service/volumeservice/types.go \
        hivepaas_app/service/volumeservice/volumeserviceimpl/app_mount_describe.go \
        hivepaas_app/service/volumeservice/volumeserviceimpl/app_mount_describe_test.go
git commit -F - <<'EOF'
fix(storage): name another app's directory only in its own environment

A directory in a volume shared wider than one environment was named after
whichever app had its key here, so a mount into staging/postgres read as
prod's postgres. The description now also reports the volume each mount
reaches, which export needs to write a mount the way it was built.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 2: The storage block can say what export needs

**Files:**
- Modify: `hivepaas_app/service/specservice/specmodel/deployment.go` (`Storage`, `Mount`)
- Test: `hivepaas_app/service/specservice/specmodel/buildable_test.go`
  (`TestCheckBuildableRefusesTheRest`)

**Interfaces:**
- Produces: `specmodel.Storage.DockerMounts map[string]Mount` (`yaml:"dockerMounts,omitempty"`)
  and `specmodel.Mount.External *ExternalRef` (`yaml:"external,omitempty"`).

- [ ] **Step 1: Write the failing tests**

In `buildable_test.go`, add two cases to the table in `TestCheckBuildableRefusesTheRest`:

```go
		"docker mounts": {"deployment:\n  storage:\n    dockerMounts:\n      /tmp: {type: tmpfs}\n",
			"deployment.storage.dockerMounts"},
		"a volume named outside the document": {"deployment:\n  storage:\n    mounts:\n" +
			"      /data: {type: volume, external: {type: cluster-volume, name: v}}\n",
			"deployment.storage.mounts./data.external"},
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./hivepaas_app/service/specservice/specmodel/ -run TestCheckBuildableRefusesTheRest -v`

Expected: FAIL in both new cases - yaml.v3 ignores a key the struct lacks, so the document
decodes clean and `CheckBuildable` returns nil where an `ErrSpecBlockUnsupported` is expected.

- [ ] **Step 3: Add the fields**

In `hivepaas_app/service/specservice/specmodel/deployment.go`, `Storage` becomes:

```go
// Storage keys mounts by their target path rather than by list position.
//
// A positional identity would be adequate while a spec carries only
// configuration - a reordered mount list re-imports the same either way. It
// stops being adequate as soon as anything outside the spec points at a mount,
// which is what a snapshot does when it pins volume data to one. Data pinned to
// mounts[0], a reorder, and a restore is data attached to the wrong volume,
// with no error and no warning.
//
// Two mounts at one target is meaningless, which makes the target a key across
// both maps - but nothing upstream validates it, so export checks and refuses.
type Storage struct {
	// Mounts are the mounts HivePaaS manages. Each reaches a volume's directory
	// for an app of this environment - the app's own, or another's named by
	// SourceApp - and is what a template writes and the storage screen edits.
	Mounts map[string]Mount `yaml:"mounts,omitempty"`
	// DockerMounts are every other mount, as Docker holds it: a tmpfs, a bind to
	// a host path no volume accounts for, a volume mounted whole. Their sources
	// are what this installation calls them, so import passes them on as they
	// are.
	DockerMounts map[string]Mount `yaml:"dockerMounts,omitempty"`
}
```

and `Mount` becomes:

```go
// Mount carries no Target: it is the map key. It carries no Key either, since
// that was a sha256 of the other three fields.
type Mount struct {
	Type mount.Type `yaml:"type"`
	// Source names the volume of a mount in Mounts - a template gives the id of a
	// cluster-volume setting, export gives the volume's scope path. In
	// DockerMounts it is what Docker holds: a volume's name, a host path.
	Source string `yaml:"source,omitempty"`
	// External stands in for Source when the volume of a mount in Mounts lies
	// outside the export, such as a volume sync discovered.
	External       *ExternalRef      `yaml:"external,omitempty"`
	ReadOnly       bool              `yaml:"readOnly,omitempty"`
	Consistency    mount.Consistency `yaml:"consistency,omitempty"`
	BindOptions    *BindOptions      `yaml:"bindOptions,omitempty"`
	VolumeOptions  *VolumeOptions    `yaml:"volumeOptions,omitempty"`
	ClusterOptions *VolumeOptions    `yaml:"clusterOptions,omitempty"`
	TmpfsOptions   *TmpfsOptions     `yaml:"tmpfsOptions,omitempty"`
	SourceApp      *MountSourceApp   `yaml:"sourceApp,omitempty"`
}
```

`CheckBuildable` needs no change: `checkStorage` allows only `mounts`, and a mount only `type`,
`source`, `readOnly`, `volumeOptions` and `sourceApp`, so both new fields are refused by name.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./hivepaas_app/service/specservice/... -v -run 'TestCheckBuildable|TestBuild|TestExport'`

Expected: PASS.

- [ ] **Step 5: Lint and commit**

```bash
go build ./... && golangci-lint run ./...
git add hivepaas_app/service/specservice/specmodel/deployment.go \
        hivepaas_app/service/specservice/specmodel/buildable_test.go
git commit -F - <<'EOF'
feat(spec): let a storage block hold Docker mounts and external volumes

Managed mounts keep the form a template writes. Everything else travels
in dockerMounts as Docker holds it, and a managed mount whose volume the
export does not hold names it by an external reference. A template may
use neither.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 3: Map an app's mounts into the storage block

**Files:**
- Create: `hivepaas_app/service/specservice/specserviceimpl/storage_map.go`
- Create: `hivepaas_app/service/specservice/specserviceimpl/storage_map_test.go`
- Modify: `hivepaas_app/service/specservice/specserviceimpl/swarm_map.go` (`mapSwarmService`;
  delete `mapStorage`)
- Modify: `hivepaas_app/service/specservice/specserviceimpl/swarm_map_test.go` (delete
  `TestMapSwarmServiceKeysMountsByTarget` and `TestMapSwarmServiceRefusesDuplicateMountTargets`,
  which move to `storage_map_test.go`)

**Interfaces:**
- Consumes: `mapMount(*mount.Mount) specmodel.Mount` (existing, `swarm_map.go`);
  `dockerhelper.GetShmMount(*swarm.TaskSpec) *mount.Mount`; `volumeservice.AppMountDesc` from
  Task 1; the fields from Task 2.
- Produces: `type volumeRef func(volumeID string) (path string, external *specmodel.ExternalRef)`
  and `mapAppStorage(task *swarm.TaskSpec, descs []*volumeservice.AppMountDesc, ref volumeRef)
  (*specmodel.Storage, error)`, which Task 5 calls.

- [ ] **Step 1: Write the failing tests**

Create `hivepaas_app/service/specservice/specserviceimpl/storage_map_test.go`:

```go
package specserviceimpl

import (
	"maps"
	"slices"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

func storageTask(mounts ...mount.Mount) *swarm.TaskSpec {
	return &swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{Mounts: mounts}}
}

func fixedVolumeRef(paths map[string]string, external map[string]*specmodel.ExternalRef) volumeRef {
	return func(volumeID string) (string, *specmodel.ExternalRef) {
		return paths[volumeID], external[volumeID]
	}
}

// A mount into the app's own directory is written the way the storage screen
// writes it, whatever form Docker was handed: a managed local volume reaches
// Docker as a bind.
func TestMapAppStorageWritesTheAppsOwnDirectoryAsAVolume(t *testing.T) {
	task := storageTask(mount.Mount{
		Type: mount.TypeBind, Source: "/srv/data/prod/web/uploads", Target: "/data", ReadOnly: true,
	})
	descs := []*volumeservice.AppMountDesc{{AppKey: "web", Own: true, Subpath: "uploads", VolumeID: "vol-1"}}

	out, err := mapAppStorage(task, descs, fixedVolumeRef(map[string]string{"vol-1": "projects/shop/volumes/data"}, nil))

	assert.NoError(t, err)
	assert.Equal(t, map[string]specmodel.Mount{
		"/data": {
			Type: mount.TypeVolume, Source: "projects/shop/volumes/data", ReadOnly: true,
			VolumeOptions: &specmodel.VolumeOptions{Subpath: "uploads"},
		},
	}, out.Mounts)
	assert.Empty(t, out.DockerMounts)
}

// Another app's directory names that app, and whether it may be changed is what
// Write says - the builder derives the mount's read-only flag from it.
func TestMapAppStorageNamesAnotherAppsDirectory(t *testing.T) {
	task := storageTask(mount.Mount{
		Type: mount.TypeVolume, Source: "hp-vol-1", Target: "/pg", ReadOnly: true,
		VolumeOptions: &mount.VolumeOptions{Subpath: "prod/postgres/pgdata", NoCopy: true},
	})
	descs := []*volumeservice.AppMountDesc{{AppKey: "postgres", Subpath: "pgdata", VolumeID: "vol-1"}}

	out, err := mapAppStorage(task, descs, fixedVolumeRef(map[string]string{"vol-1": "projects/shop/volumes/data"}, nil))

	assert.NoError(t, err)
	assert.Equal(t, specmodel.Mount{
		Type: mount.TypeVolume, Source: "projects/shop/volumes/data",
		VolumeOptions: &specmodel.VolumeOptions{Subpath: "pgdata", NoCopy: true},
		SourceApp:     &specmodel.MountSourceApp{App: "postgres", Write: false},
	}, out.Mounts["/pg"])
}

// A volume the export does not hold is named by an external reference.
func TestMapAppStorageNamesAVolumeOutsideTheExport(t *testing.T) {
	task := storageTask(mount.Mount{
		Type: mount.TypeVolume, Source: "hp-shared", Target: "/shared",
		VolumeOptions: &mount.VolumeOptions{Subpath: "shop/prod/web"},
	})
	descs := []*volumeservice.AppMountDesc{{AppKey: "web", Own: true, VolumeID: "gvol-1"}}
	shared := &specmodel.ExternalRef{Type: "cluster-volume", Name: "shared", ID: "gvol-1"}

	out, err := mapAppStorage(task, descs, fixedVolumeRef(nil, map[string]*specmodel.ExternalRef{"gvol-1": shared}))

	assert.NoError(t, err)
	assert.Equal(t, specmodel.Mount{Type: mount.TypeVolume, External: shared}, out.Mounts["/shared"])
}

// Every other mount is kept as Docker holds it, and the shared-memory mount is
// kept by neither map: resources.memory.shmSize carries it.
func TestMapAppStorageKeepsOtherMountsAsDockerHoldsThem(t *testing.T) {
	task := storageTask(
		mount.Mount{Type: mount.TypeTmpfs, Target: "/dev/shm", TmpfsOptions: &mount.TmpfsOptions{SizeBytes: 64 << 20}},
		mount.Mount{Type: mount.TypeBind, Source: "/etc/localtime", Target: "/etc/localtime", ReadOnly: true},
		mount.Mount{Type: mount.TypeVolume, Source: "hp-vol-1", Target: "/whole"},
	)
	descs := []*volumeservice.AppMountDesc{{}, {}, {}}

	out, err := mapAppStorage(task, descs, fixedVolumeRef(nil, nil))

	assert.NoError(t, err)
	assert.Empty(t, out.Mounts)
	assert.Equal(t, []string{"/etc/localtime", "/whole"}, slices.Sorted(maps.Keys(out.DockerMounts)))
	assert.Equal(t, specmodel.Mount{Type: mount.TypeBind, Source: "/etc/localtime", ReadOnly: true},
		out.DockerMounts["/etc/localtime"])
}

// A volume nothing can name any more - its setting is gone - leaves the mount as
// Docker holds it rather than losing it.
func TestMapAppStorageKeepsAMountWhoseVolumeCannotBeNamed(t *testing.T) {
	task := storageTask(mount.Mount{
		Type: mount.TypeVolume, Source: "hp-gone", Target: "/data",
		VolumeOptions: &mount.VolumeOptions{Subpath: "prod/web"},
	})
	descs := []*volumeservice.AppMountDesc{{AppKey: "web", Own: true, VolumeID: "gone"}}

	out, err := mapAppStorage(task, descs, fixedVolumeRef(nil, nil))

	assert.NoError(t, err)
	assert.Empty(t, out.Mounts)
	assert.Equal(t, "hp-gone", out.DockerMounts["/data"].Source)
}

// Nothing upstream validates that two mounts do not share a target, so this is
// the only thing holding the invariant a future snapshot pins data to.
func TestMapAppStorageRefusesDuplicateMountTargets(t *testing.T) {
	task := storageTask(
		mount.Mount{Type: mount.TypeVolume, Source: "vol_1", Target: "/data"},
		mount.Mount{Type: mount.TypeBind, Source: "/srv", Target: "/data"},
	)

	_, err := mapAppStorage(task, []*volumeservice.AppMountDesc{{}, {}}, fixedVolumeRef(nil, nil))

	assert.ErrorIs(t, err, hperrors.ErrSpecMountTargetDuplicated)
}

func TestMapAppStorageIsNilWithoutMounts(t *testing.T) {
	out, err := mapAppStorage(storageTask(), nil, fixedVolumeRef(nil, nil))
	assert.NoError(t, err)
	assert.Nil(t, out)

	out, err = mapAppStorage(&swarm.TaskSpec{}, nil, fixedVolumeRef(nil, nil))
	assert.NoError(t, err)
	assert.Nil(t, out)
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run TestMapAppStorage -v`

Expected: the package does not compile - `undefined: volumeRef` and `undefined: mapAppStorage`.

- [ ] **Step 3: Write the mapping**

Create `hivepaas_app/service/specservice/specserviceimpl/storage_map.go`:

```go
package specserviceimpl

import (
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

// volumeRef is how the bundle names the volume a managed mount reaches: its
// scope path when the export holds the volume, an external reference when it
// does not. Both empty means the volume cannot be named at all.
type volumeRef func(volumeID string) (path string, external *specmodel.ExternalRef)

// mapAppStorage turns an app's mounts into its storage block.
//
// A mount into a volume's directory for an app of this environment is written
// the way a template and the storage screen write it: the volume, the path below
// that app's directory, and the app when it is not this one. That is the form
// import can build again anywhere - Docker's own form carries a host path or a
// volume name that means nothing on another installation. Every other mount is
// kept as Docker holds it. The shared-memory mount is neither:
// resources.memory.shmSize carries it.
//
// descs is volumeservice.DescribeAppMounts' answer for the same mounts, in the
// same order.
func mapAppStorage(
	task *swarm.TaskSpec,
	descs []*volumeservice.AppMountDesc,
	ref volumeRef,
) (*specmodel.Storage, error) {
	if task.ContainerSpec == nil || len(task.ContainerSpec.Mounts) == 0 {
		return nil, nil
	}
	mounts := task.ContainerSpec.Mounts
	shm := dockerhelper.GetShmMount(task)

	out := &specmodel.Storage{}
	seen := make(map[string]bool, len(mounts))
	for i := range mounts {
		m := &mounts[i]
		if shm != nil && m.Type == shm.Type && m.Target == shm.Target {
			continue
		}
		if seen[m.Target] {
			return nil, hperrors.Wrap(hperrors.ErrSpecMountTargetDuplicated).WithParam("Target", m.Target)
		}
		seen[m.Target] = true

		var desc *volumeservice.AppMountDesc
		if i < len(descs) {
			desc = descs[i]
		}
		if managed, ok := mapManagedMount(m, desc, ref); ok {
			out.Mounts = withMount(out.Mounts, m.Target, managed)
			continue
		}
		out.DockerMounts = withMount(out.DockerMounts, m.Target, mapMount(m))
	}
	if len(out.Mounts) == 0 && len(out.DockerMounts) == 0 {
		return nil, nil
	}
	return out, nil
}

// mapManagedMount writes a mount the way the storage screen writes it, when it
// is one: a directory an app of this environment has in a volume the bundle can
// name.
func mapManagedMount(
	m *mount.Mount,
	desc *volumeservice.AppMountDesc,
	ref volumeRef,
) (specmodel.Mount, bool) {
	if desc == nil || desc.AppKey == "" || desc.VolumeID == "" {
		return specmodel.Mount{}, false
	}
	path, external := ref(desc.VolumeID)
	if path == "" && external == nil {
		return specmodel.Mount{}, false
	}

	// A managed local volume reaches Docker as a bind, but what was asked for is
	// the volume, and building that again makes the same bind.
	out := specmodel.Mount{
		Type:        mount.TypeVolume,
		Source:      path,
		External:    external,
		Consistency: m.Consistency,
	}
	noCopy := m.VolumeOptions != nil && m.VolumeOptions.NoCopy
	var opts *specmodel.VolumeOptions
	if desc.Subpath != "" || noCopy {
		opts = &specmodel.VolumeOptions{Subpath: desc.Subpath, NoCopy: noCopy}
	}
	if m.Type == mount.TypeCluster {
		out.Type, out.ClusterOptions = mount.TypeCluster, opts
	} else {
		out.VolumeOptions = opts
	}

	if desc.Own {
		out.ReadOnly = m.ReadOnly
	} else {
		// Whether this app may change the other's files is what Write says; the
		// builder derives the mount's read-only flag from it, so it is not
		// written twice.
		out.SourceApp = &specmodel.MountSourceApp{App: desc.AppKey, Write: !m.ReadOnly}
	}
	return out, true
}

func withMount(mounts map[string]specmodel.Mount, target string, m specmodel.Mount) map[string]specmodel.Mount {
	if mounts == nil {
		mounts = map[string]specmodel.Mount{}
	}
	mounts[target] = m
	return mounts
}
```

- [ ] **Step 4: Take storage out of mapSwarmService**

The storage block now needs the database to read (Task 5), so `mapSwarmService` stops setting
it - and with it, the only thing it could fail on. In
`hivepaas_app/service/specservice/specserviceimpl/swarm_map.go`, `mapSwarmService` becomes:

```go
// mapSwarmService turns a live Swarm service into the declarative part of it,
// storage aside: which volume a mount reaches is a question for the database,
// and mapAppStorage answers it.
//
// netNames maps Docker network id to name; an id with no entry is kept as-is,
// so an unresolved attachment is reported by import rather than lost here.
func mapSwarmService(
	svc *swarm.Service,
	netNames map[string]string,
) *specmodel.Deployment {
	if svc == nil {
		return nil
	}
	spec := &svc.Spec
	task := &spec.TaskTemplate

	out := &specmodel.Deployment{
		Resources: mapResources(task),
		Networks:  mapNetworks(task, spec.EndpointSpec, netNames),
		Service:   mapService(spec, task),
	}
	if cs := task.ContainerSpec; cs != nil {
		out.Container = mapContainer(cs, task, spec.Labels)
	}
	return out
}
```

Delete the `mapStorage` function and the now unused `hperrors` import; keep `mapMount`, which
`mapAppStorage` uses.

In `walk.go`, `addSwarmBlocks` calls it without an error:

```go
	mapped := mapSwarmService(svc, netNames)
	if mapped == nil {
		return nil
	}
```

In `swarm_map_test.go`, delete `TestMapSwarmServiceKeysMountsByTarget` and
`TestMapSwarmServiceRefusesDuplicateMountTargets` - `storage_map_test.go` covers both now - and
drop the error from every remaining call:

```bash
perl -0pi -e 's/out, err := mapSwarmService\((.*?)\)\n\tassert\.NoError\(t, err\)/out := mapSwarmService($1)/gs' \
  hivepaas_app/service/specservice/specserviceimpl/swarm_map_test.go
grep -n "err" hivepaas_app/service/specservice/specserviceimpl/swarm_map_test.go
```

Expected: the `grep` prints nothing.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./hivepaas_app/service/specservice/... -v -run 'TestMapAppStorage|TestMapSwarmService|TestExport'`

Expected: PASS. The export tests still pass: until Task 5, an export has no storage block.

- [ ] **Step 6: Lint and commit**

```bash
go build ./... && golangci-lint run ./...
git add hivepaas_app/service/specservice/specserviceimpl/storage_map.go \
        hivepaas_app/service/specservice/specserviceimpl/storage_map_test.go \
        hivepaas_app/service/specservice/specserviceimpl/swarm_map.go \
        hivepaas_app/service/specservice/specserviceimpl/swarm_map_test.go
git commit -F - <<'EOF'
feat(spec): map an app's mounts the way they were built

A mount into an app's directory in a volume is written as a template and
the storage screen write it - the volume, the path below the app's
directory, and the owning app when it is another - which import can
build again on any installation. Every other mount is kept as Docker
holds it, and the shared-memory mount is left to memory.shmSize.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 4: A reference that leaves the export becomes an external reference

**Files:**
- Modify: `hivepaas_app/service/specservice/specserviceimpl/service.go` (`settingsByIDLoader`,
  `service.loadByIDs`, `New`)
- Modify: `hivepaas_app/service/specservice/specserviceimpl/walk.go` (`buildBundle`; new
  `indexExternalRefs`, `externalRef`, `loadByIDsFromRepo`)
- Modify: `hivepaas_app/service/specservice/specserviceimpl/export_test.go` (`fakeProjectRepo.GetByID`,
  the fixture's `loadByIDs`, `runExportAt`)
- Create: `hivepaas_app/service/specservice/specserviceimpl/export_refs_test.go`

**Interfaces:**
- Consumes: `refIndex.addExternal(*entity.Setting)` and `replaceExternalRefs` (existing,
  `assemble.go`); `readDoc` (`export_ids_test.go`).
- Produces: `type settingsByIDLoader func(ctx, db, ids []string) ([]*entity.Setting, error)`
  and the `service.loadByIDs` field, which Task 5 uses for volumes; test helper
  `runExportAt(t, scope, mode, passphrase)`.

- [ ] **Step 1: Let the fixture export a single project**

In `export_test.go`, add to `fakeProjectRepo`:

```go
func (f *fakeProjectRepo) GetByID(
	_ context.Context, _ database.IDB, id string, _ ...bunex.SelectQueryOption,
) (*entity.Project, error) {
	for _, project := range f.projects {
		if project.ID == id {
			return project, nil
		}
	}
	return nil, notFoundError{}
}
```

In `exportFixture`, after `impl.loadOwned = ...`, the fixture answers loads by id from the same
settings:

```go
	impl.loadByIDs = func(_ context.Context, _ database.IDB, ids []string) ([]*entity.Setting, error) {
		var out []*entity.Setting
		for _, setting := range all {
			if slices.Contains(ids, setting.ID) {
				out = append(out, setting)
			}
		}
		return out, nil
	}
```

(add `"slices"` to the imports), and `runExport` becomes a special case of a new helper:

```go
func runExport(t *testing.T, mode specmodel.SecretsMode, passphrase string) (string, *specmodel.Report) {
	t.Helper()
	return runExportAt(t, entity.NewObjectScopeGlobal(), mode, passphrase)
}

func runExportAt(
	t *testing.T, scope *entity.ObjectScope, mode specmodel.SecretsMode, passphrase string,
) (string, *specmodel.Report) {
	t.Helper()
	svc := exportFixture(t)
	dir := t.TempDir()

	resp, err := svc.Export(context.Background(), nil, &specservice.ExportReq{
		Scope:       scope,
		SecretsMode: mode,
		Passphrase:  passphrase,
		WorkDir:     dir,
	})
	assert.NoError(t, err)
	assert.FileExists(t, resp.Path)
	return resp.Path, resp.Report
}
```

- [ ] **Step 2: Write the failing test**

Create `hivepaas_app/service/specservice/specserviceimpl/export_refs_test.go`:

```go
package specserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// A project exported alone does not hold the global certificate its app uses.
// The reference becomes an external one: the id finds the certificate again on
// the installation that exported it, the type, name and kind anywhere else.
func TestExportWritesAReferenceOutsideTheExportAsExternal(t *testing.T) {
	path, _ := runExportAt(t, entity.NewObjectScopeProject("p1"), specmodel.SecretsModeOmit, "")

	env := &specmodel.EnvDoc{}
	readDoc(t, path, "projects/project_a/envs/dev.yaml", env)
	routing := env.Apps["backend"].Settings["routing"].(map[string]any)
	domain := routing["domains"].([]any)[0].(map[string]any)

	assert.Equal(t, map[string]any{"id": map[string]any{"external": map[string]any{
		"type": "ssl-cert", "name": "localhost", "kind": "self-signed", "id": "cert_1",
	}}}, domain["sslCert"])
}

// The whole installation holds the certificate, so the same reference is a path.
func TestExportWritesAReferenceInsideTheExportAsAPath(t *testing.T) {
	path, _ := runExport(t, specmodel.SecretsModeOmit, "")

	env := &specmodel.EnvDoc{}
	readDoc(t, path, "projects/project_a/envs/dev.yaml", env)
	routing := env.Apps["backend"].Settings["routing"].(map[string]any)
	domain := routing["domains"].([]any)[0].(map[string]any)

	assert.Equal(t, map[string]any{"id": "global/sslCerts/localhost"}, domain["sslCert"])
}
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run 'TestExportWritesAReference' -v`

Expected: the package does not compile - `impl.loadByIDs undefined`.

- [ ] **Step 4: Add the seam**

In `hivepaas_app/service/specservice/specserviceimpl/service.go`, below `settingLoader`:

```go
// settingsByIDLoader loads settings by id, whatever their scope. Export asks it
// for the settings a reference reaches outside what is exported. It is a seam
// for the reason settingLoader is: a test double cannot read bunex options.
type settingsByIDLoader func(ctx context.Context, db database.IDB, ids []string) ([]*entity.Setting, error)
```

`service` gains a field after `loadOwned`:

```go
	loadOwned settingLoader
	loadByIDs settingsByIDLoader
```

and `New` sets it after `svc.loadOwned = svc.loadOwnedFromRepo`:

```go
	svc.loadByIDs = svc.loadByIDsFromRepo
```

- [ ] **Step 5: Register external references before any document is written**

In `hivepaas_app/service/specservice/specserviceimpl/walk.go`, in `buildBundle`, between pass one
and pass two:

```go
	// Pass one and a half: a reference to a setting the export does not hold - a
	// project exported alone whose app uses a global certificate - is written as
	// an external reference, which says what finds its target again elsewhere.
	if err = s.indexExternalRefs(ctx, db, tree, index); err != nil {
		return nil, hperrors.Wrap(err)
	}
```

Below `loadOwnedFromRepo`, add:

```go
func (s *service) loadByIDsFromRepo(
	ctx context.Context,
	db database.IDB,
	ids []string,
) ([]*entity.Setting, error) {
	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhereIn("setting.id IN (?)", ids...),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return settings, nil
}

// indexExternalRefs registers every setting an exported setting references but
// the export does not hold. replaceExternalRefs then writes each reference to it
// as an external one. A reference to a setting that no longer exists is left as
// the id it is, for import to report.
func (s *service) indexExternalRefs(
	ctx context.Context,
	db database.IDB,
	tree *exportTree,
	index *refIndex,
) error {
	var ids []string
	seen := map[string]bool{}
	for _, unit := range tree.units {
		for _, setting := range unit.settings {
			refs, err := setting.GetRefObjectIDs()
			if err != nil {
				return hperrors.Wrap(err)
			}
			if refs == nil {
				continue
			}
			for _, id := range refs.RefSettingIDs {
				if id == "" || seen[id] || index.paths[id] != "" {
					continue
				}
				seen[id] = true
				ids = append(ids, id)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	settings, err := s.loadByIDs(ctx, db, ids)
	if err != nil {
		return hperrors.Wrap(err)
	}
	for _, setting := range settings {
		index.addExternal(setting)
	}
	return nil
}
```

- [ ] **Step 6: Run the tests to verify they pass**

Run: `go test ./hivepaas_app/service/specservice/... -v -run 'TestExport'`

Expected: PASS, the determinism test included.

- [ ] **Step 7: Lint and commit**

```bash
go build ./... && golangci-lint run ./...
git add hivepaas_app/service/specservice/specserviceimpl/service.go \
        hivepaas_app/service/specservice/specserviceimpl/walk.go \
        hivepaas_app/service/specservice/specserviceimpl/export_test.go \
        hivepaas_app/service/specservice/specserviceimpl/export_refs_test.go
git commit -F - <<'EOF'
fix(spec): write a reference that leaves the export as an external one

The external form was designed and never wired: nothing registered the
settings a reference reached outside the export, so exporting a project
alone left its apps pointing at the raw id of a global certificate.
Those settings are now loaded before any document is written, and each
reference to one says what finds it again on another installation.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 5: Export writes an app's storage the way it was built

**Files:**
- Modify: `hivepaas_app/service/specservice/specserviceimpl/walk.go` (`buildBundle`, `writeDocs`,
  `writeEnvDoc`, `buildAppDoc`, `addSwarmBlocks`; new `mapStorageOf`)
- Modify: `hivepaas_app/service/specservice/specserviceimpl/export_test.go` (fixture: two volume
  settings, a volume service)
- Modify: `hivepaas_app/service/specservice/specserviceimpl/swarm_map_test.go` (`testService`
  gains two mounts)
- Test: `hivepaas_app/service/specservice/specserviceimpl/export_refs_test.go`

**Interfaces:**
- Consumes: `mapAppStorage` and `volumeRef` (Task 3); `service.loadByIDs` (Task 4);
  `volumeservice.Service.DescribeAppMounts(ctx, db, app, mounts)`.
- Produces: `(*service).externalRef(ctx, db, index, id) (*specmodel.ExternalRef, error)` and
  `(*service).mapStorageOf(ctx, db, app *entity.App, svc *swarm.Service, index *refIndex)
  (*specmodel.Storage, error)`.

- [ ] **Step 1: Give the fixture volumes and a volume service**

In `swarm_map_test.go`, `testService`'s mounts become:

```go
					Mounts: []mount.Mount{
						{Type: mount.TypeVolume, Source: "vol_1", Target: "/var/lib/postgresql/data"},
						{Type: mount.TypeBind, Source: "/srv/conf", Target: "/etc/app/config"},
						{
							Type: mount.TypeVolume, Source: "hp-shared", Target: "/shared",
							VolumeOptions: &mount.VolumeOptions{Subpath: "project_a/dev/backend/cache"},
						},
						{
							Type: mount.TypeTmpfs, Target: "/dev/shm",
							TmpfsOptions: &mount.TmpfsOptions{SizeBytes: 64 << 20},
						},
					},
```

In `export_test.go`, add the volume service double after `fakeClusterService`:

```go
// fakeExportVolumeService answers DescribeAppMounts from a table keyed by mount
// target, which is all export asks of volumeservice.
type fakeExportVolumeService struct {
	volumeservice.Service
	descs map[string]*volumeservice.AppMountDesc
}

func (f *fakeExportVolumeService) DescribeAppMounts(
	_ context.Context, _ database.IDB, _ *entity.App, mounts []mount.Mount,
) ([]*volumeservice.AppMountDesc, error) {
	out := make([]*volumeservice.AppMountDesc, len(mounts))
	for i := range mounts {
		if out[i] = f.descs[mounts[i].Target]; out[i] == nil {
			out[i] = &volumeservice.AppMountDesc{}
		}
	}
	return out, nil
}
```

(imports: `"github.com/moby/moby/api/types/mount"` and
`"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"`).

In `exportFixture`, before `settingRepo := &fakeSettingRepo{}`:

```go
	// The project's own volume, which the export holds, and a volume sync
	// discovered at global scope, which it never does.
	projectVolume := &entity.Setting{
		ID: "vol_setting_1", Type: base.SettingTypeClusterVolume, Scope: base.ObjectScopeProject,
		ObjectID: "p1", Name: "default", Status: base.SettingStatusActive,
	}
	assert.NoError(t, projectVolume.SetData(&entity.ClusterVolume{Managed: true}))
	sharedVolume := &entity.Setting{
		ID: "gvol_1", Type: base.SettingTypeClusterVolume, Scope: base.ObjectScopeGlobal,
		Name: "shared", Status: base.SettingStatusActive,
	}
	assert.NoError(t, sharedVolume.SetData(&entity.ClusterVolume{}))
```

`all` becomes `[]*entity.Setting{cert, apiKey, routing, secret, projectVolume, sharedVolume}`,
and the last argument to `New` - the volume service, `nil` until now - becomes:

```go
		&fakeExportVolumeService{descs: map[string]*volumeservice.AppMountDesc{
			"/var/lib/postgresql/data": {AppKey: "backend", Own: true, Subpath: "data", VolumeID: "vol_setting_1"},
			"/shared":                  {AppKey: "backend", Own: true, Subpath: "cache", VolumeID: "gvol_1"},
		}},
```

- [ ] **Step 2: Write the failing tests**

Append to `export_refs_test.go` (imports gain `"github.com/moby/moby/api/types/mount"` and
`"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"`):

```go
func exportedStorage(t *testing.T) (*specmodel.Storage, *specmodel.Resources) {
	t.Helper()
	path, _ := runExport(t, specmodel.SecretsModeOmit, "")
	env := &specmodel.EnvDoc{}
	readDoc(t, path, "projects/project_a/envs/dev.yaml", env)
	deployment := env.Apps["backend"].Deployment
	if !assert.NotNil(t, deployment) || !assert.NotNil(t, deployment.Storage) {
		t.FailNow()
	}
	return deployment.Storage, deployment.Resources
}

// A mount into the app's directory is written the way the storage screen
// writes it: the volume by its path in the bundle, the directory below the
// app's own.
func TestExportWritesAManagedMountAsTheStorageScreenDoes(t *testing.T) {
	storage, _ := exportedStorage(t)

	assert.Equal(t, specmodel.Mount{
		Type: mount.TypeVolume, Source: "projects/project_a/volumes/default",
		VolumeOptions: &specmodel.VolumeOptions{Subpath: "data"},
	}, storage.Mounts["/var/lib/postgresql/data"])
}

// A volume sync discovered is never exported, so a mount into it names the
// volume by an external reference.
func TestExportNamesAVolumeOutsideTheExportByReference(t *testing.T) {
	storage, _ := exportedStorage(t)

	assert.Equal(t, specmodel.Mount{
		Type:          mount.TypeVolume,
		External:      &specmodel.ExternalRef{Type: "cluster-volume", Name: "shared", ID: "gvol_1"},
		VolumeOptions: &specmodel.VolumeOptions{Subpath: "cache"},
	}, storage.Mounts["/shared"])
}

// Every other mount is kept as Docker holds it, and the shared-memory mount is
// carried by resources.memory.shmSize alone.
func TestExportKeepsOtherMountsAsDockerHoldsThem(t *testing.T) {
	storage, resources := exportedStorage(t)

	assert.Equal(t, specmodel.Mount{Type: mount.TypeBind, Source: "/srv/conf"},
		storage.DockerMounts["/etc/app/config"])
	assert.NotContains(t, storage.Mounts, "/dev/shm")
	assert.NotContains(t, storage.DockerMounts, "/dev/shm")
	if assert.NotNil(t, resources) && assert.NotNil(t, resources.Memory) && assert.NotNil(t, resources.Memory.ShmSize) {
		assert.Equal(t, unit.DataSize(64<<20), *resources.Memory.ShmSize)
	}
}
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `go test ./hivepaas_app/service/specservice/specserviceimpl/ -run 'TestExportWritesAManagedMount|TestExportNamesAVolume|TestExportKeepsOtherMounts' -v`

Expected: FAIL - `deployment.Storage` is nil (`Should not be nil`), since Task 3 took storage
out of `mapSwarmService` and nothing puts it back yet.

- [ ] **Step 4: Thread the database to the app documents**

In `walk.go`, `buildBundle` calls `s.writeDocs(ctx, db, req, tree, netNames, index, bundle)`.
`writeDocs`, `writeEnvDoc`, `buildAppDoc` and `addSwarmBlocks` each gain `db database.IDB` as
their second parameter, and pass it on:

```go
func (s *service) writeDocs(
	ctx context.Context,
	db database.IDB,
	req *specservice.ExportReq,
	tree *exportTree,
	netNames map[string]string,
	index *refIndex,
	bundle *specmodel.Bundle,
) error {
```

```go
			if err := s.writeEnvDoc(ctx, db, req, projUnit, envUnit, netNames, index, bundle); err != nil {
```

```go
func (s *service) writeEnvDoc(
	ctx context.Context,
	db database.IDB,
	req *specservice.ExportReq,
	projUnit *projectUnit,
	env *envUnit,
	netNames map[string]string,
	index *refIndex,
	bundle *specmodel.Bundle,
) error {
	appDocs := map[string]*specmodel.AppDoc{}
	for _, app := range env.apps {
		// Where an app's directory is inside a volume depends on its project and
		// environment, which the app row alone does not carry.
		app.app.Project, app.app.ProjectEnv = projUnit.project, env.env
		doc, err := s.buildAppDoc(ctx, db, req, app, netNames, index, bundle.Report)
```

```go
func (s *service) buildAppDoc(
	ctx context.Context,
	db database.IDB,
	req *specservice.ExportReq,
	app *appUnit,
	netNames map[string]string,
	index *refIndex,
	report *specmodel.Report,
) (*specmodel.AppDoc, error) {
```

```go
	if err = s.addSwarmBlocks(ctx, db, app.app, app.unit.path, netNames, index, deployment, report); err != nil {
```

- [ ] **Step 5: Map the storage in addSwarmBlocks**

Below `indexExternalRefs` in `walk.go`, add the lookup a volume outside the export needs - one
at a time, since which volumes are mounted is learned only from the services, after the
external references of settings were registered:

```go
// externalRef is the external reference for one setting the export does not
// hold, loaded the first time it is asked for. It is nil when no such setting
// exists.
func (s *service) externalRef(
	ctx context.Context,
	db database.IDB,
	index *refIndex,
	id string,
) (*specmodel.ExternalRef, error) {
	if ref := index.external[id]; ref != nil {
		return ref, nil
	}
	settings, err := s.loadByIDs(ctx, db, []string{id})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, setting := range settings {
		index.addExternal(setting)
	}
	return index.external[id], nil
}
```

`addSwarmBlocks` gains `db database.IDB` second and `index *refIndex` after `netNames`, and
ends:

```go
	mapped := mapSwarmService(svc, netNames)
	if mapped == nil {
		return nil
	}
	storage, err := s.mapStorageOf(ctx, db, app, svc, index)
	if err != nil {
		return hperrors.Wrap(err)
	}
	deployment.Container = mapped.Container
	deployment.Resources = mapped.Resources
	deployment.Storage = storage
	deployment.Networks = mapped.Networks
	deployment.Service = mapped.Service
	return nil
}

// mapStorageOf reads an app's mounts back into what built them. volumeservice
// says which volume and whose directory each one reaches - the same answer the
// storage screen shows - and each volume is named the way the bundle can: by
// its path when the export holds it, by an external reference when it does not.
func (s *service) mapStorageOf(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	svc *swarm.Service,
	index *refIndex,
) (*specmodel.Storage, error) {
	task := &svc.Spec.TaskTemplate
	if task.ContainerSpec == nil || len(task.ContainerSpec.Mounts) == 0 {
		return nil, nil
	}
	descs, err := s.volumeService.DescribeAppMounts(ctx, db, app, task.ContainerSpec.Mounts)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	type volumeName struct {
		path     string
		external *specmodel.ExternalRef
	}
	names := map[string]volumeName{}
	for _, desc := range descs {
		if desc == nil || desc.VolumeID == "" {
			continue
		}
		if _, done := names[desc.VolumeID]; done {
			continue
		}
		name := volumeName{path: index.paths[desc.VolumeID]}
		if name.path == "" {
			if name.external, err = s.externalRef(ctx, db, index, desc.VolumeID); err != nil {
				return nil, hperrors.Wrap(err)
			}
		}
		names[desc.VolumeID] = name
	}
	return mapAppStorage(task, descs, func(volumeID string) (string, *specmodel.ExternalRef) {
		return names[volumeID].path, names[volumeID].external
	})
}
```

(`walk.go` imports gain `"github.com/moby/moby/api/types/swarm"`.)

- [ ] **Step 6: Run the tests to verify they pass**

Run: `go test ./hivepaas_app/service/specservice/... -v -run 'TestExport|TestMap'`

Expected: PASS - the three new tests, and every existing export test, determinism included.

- [ ] **Step 7: Lint and commit**

```bash
go build ./... && golangci-lint run ./...
git add hivepaas_app/service/specservice/specserviceimpl/walk.go \
        hivepaas_app/service/specservice/specserviceimpl/export_test.go \
        hivepaas_app/service/specservice/specserviceimpl/swarm_map_test.go \
        hivepaas_app/service/specservice/specserviceimpl/export_refs_test.go
git commit -F - <<'EOF'
feat(spec): export an app's storage the way it was built

Export asks volumeservice which volume and whose directory each mount
reaches, and writes a mount into an app's directory as the storage
screen would: the volume by its path in the bundle, or by an external
reference when the bundle does not hold it. What Docker was handed - a
bind to this host's disk, a volume's name here - no longer travels as
the description of a managed mount.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 6: Record the decisions, and verify the whole change

**Files:**
- Modify: `docs/superpowers/specs/2026-09-23-config-spec-import-design.md` (§9 storage
  paragraph, corrections list, §14)

- [ ] **Step 1: Update the import spec**

In §9, replace the paragraph that begins "**Storage is the one block export writes in a form the
builder cannot read.**" with:

```markdown
**Storage travels in the form the builder reads.** A mount into a volume's directory for an
app of its environment is exported the way a template and the storage screen write it: the
volume by its scope path - or, when the export does not hold it, by `external` - the path
below the app's directory, and `sourceApp` when the directory is another app's. Every other
mount travels in `storage.dockerMounts` exactly as Docker holds it, and the shared-memory mount
in `resources.memory.shmSize` alone. Two things a managed mount can carry do not travel: a
per-mount driver override and labels, because export cannot tell them from the volume's own,
which a mount inherits when it names none. Building the volume's mount again inherits them
again.
```

In *Corrections to the export spec*, add after item 6:

```markdown
7. **A reference that left the export was written as a raw id.** The external form was
   designed but never wired: nothing registered the settings a reference reached outside the
   export. Export now loads them first and writes each such reference as
   `{external: {type, name, kind, id}}`. A reference to an app, a project, an env or a user is
   still the source id; import resolves it through the ids the bundle carries (§4), and on
   another installation one the bundle does not carry is `REF_NOT_FOUND`.
```

In §14, item 2 (builder coverage) now starts from a storage block in the builder's form, so
replace "It starts by settling how export represents a mount (§9)." with "Storage already
travels in the builder's form (§9); the builder learns `dockerMounts`, `external` and
`clusterOptions`."

- [ ] **Step 2: The whole suite**

Run: `go build ./... && golangci-lint run ./... && go test ./...`

Expected: everything passes.

- [ ] **Step 3: Export the development database**

The volume descriptions and the id lookup run real SQL that the doubles do not. With the
development database up (`config/config.local.toml` points at it), export it through the
running development backend - Operations → Export, secrets mode *Omit* - and unpack the newest
download:

```bash
BUNDLE=$(ls -t ~/Downloads/hivepaas-spec-*.tar.gz | head -1)
rm -rf /tmp/spec-check && mkdir -p /tmp/spec-check && tar -xzf "$BUNDLE" -C /tmp/spec-check
grep -n -A6 "storage:" /tmp/spec-check/projects/*/envs/*.yaml | head -60
grep -rn "external:" /tmp/spec-check | head -20
```

Expected: each deployed app with a volume has `storage.mounts` entries whose `source` is a path
such as `projects/<key>/volumes/<name>` or whose `external` names a `cluster-volume`, and no
managed mount carries a host path; a tmpfs or a bind outside every volume sits under
`dockerMounts`; no `/dev/shm` entry appears in either map.

- [ ] **Step 4: Commit and report**

```bash
git add docs/superpowers/specs/2026-09-23-config-spec-import-design.md
git commit -F - <<'EOF'
docs(spec): record how references and storage travel

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

Say which checks ran and what they showed. If Step 3 could not be run, say so rather than
calling the task done.

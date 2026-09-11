# Volume Lazy Materialization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move a volume's specification from one node's docker daemon into its setting, stop creating the docker volume up front, and make node pinning a placement constraint swarm enforces.

**Architecture:** `entity.ClusterVolume` becomes the authoritative, write-once description of a volume (pin, driver, driver options, labels). Creation records it without calling docker. Mount specs are built from it, so docker materializes the volume correctly on whatever node a task lands. Pinned volumes additionally constrain placement so a task can only be scheduled where its data is.

**Tech Stack:** Go 1.x, moby/moby client, bun ORM, fx DI, testify (`assert`), `tiendc/gofn`, `tiendc/go-validator` (`vld`).

**Spec:** `docs/superpowers/specs/2026-09-11-volume-lazy-materialization-design.md`

## Global Constraints

- Tests: `make test` runs `./scripts/test.sh` (`go test -race ./...` with coverage). A single package: `go test ./hivepaas_app/... -run TestName -v`.
- Lint: `make lint-local` (golangci-lint v2.13.0). Import order is enforced by `gci`: stdlib, then third-party, then `github.com/hivepaas/hivepaas/...`, separated by blank lines.
- Errors: wrap with `hperrors.Wrap(err)` at every return; construct with `hperrors.NewNotFound`, `NewAlreadyExist`, `NewArgumentInvalid`, `NewUnsupported`.
- Settings data: only `SetData`/`MustSetData` write `Setting.Data`. A struct returned by `AsClusterVolume()` is **not** linked back to `Data` — always read, modify, then `SetData`.
- `Managed` means HivePaaS authored the specification and may reproduce the volume elsewhere. Volumes that arrived through discovery keep `Managed: false` and are mounted by name with nothing inferred.
- Node pinning is decided at creation and is immutable thereafter.
- Comments explain *why*, in the voice of the surrounding code. Do not narrate what the code already says.

---

### Task 1: Carry the volume specification in the setting

**Files:**
- Modify: `hivepaas_app/entity/setting_cluster_volume.go`
- Test: `hivepaas_app/entity/setting_cluster_volume_test.go` (create)

**Interfaces:**
- Consumes: nothing.
- Produces: `entity.ClusterVolume` with fields `NodeID`, `NodeLabel string`, `Managed bool`, `Driver string`, `DriverOpts map[string]string`, `Labels map[string]string`; method `func (s *ClusterVolume) IsPinned() bool`.

- [ ] **Step 1: Write the failing test**

Create `hivepaas_app/entity/setting_cluster_volume_test.go`:

```go
package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func clusterVolumeSetting(vol *ClusterVolume) *Setting {
	setting := &Setting{Type: base.SettingTypeClusterVolume}
	setting.MustSetData(vol)
	return setting
}

// The specification has to survive the trip through Data, because Data is what
// the repository writes and a field that only exists in memory is not recorded.
func TestClusterVolumeSpecRoundTrips(t *testing.T) {
	setting := clusterVolumeSetting(&ClusterVolume{
		NodeID:     "node-1",
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "none", "device": "/data/pg", "o": "bind,rw"},
		Labels:     map[string]string{"hivepaas.volume.name": "pgdata"},
	})

	parsed, err := setting.AsClusterVolume()
	assert.NoError(t, err)
	assert.Equal(t, "node-1", parsed.NodeID)
	assert.True(t, parsed.Managed)
	assert.Equal(t, "local", parsed.Driver)
	assert.Equal(t, "/data/pg", parsed.DriverOpts["device"])
	assert.Equal(t, "pgdata", parsed.Labels["hivepaas.volume.name"])
}

// A volume recorded before this field existed reads back as unmanaged, which is
// what keeps it mounted by name until a sync backfills its specification.
func TestClusterVolumeDefaultsToUnmanaged(t *testing.T) {
	setting := clusterVolumeSetting(&ClusterVolume{NodeID: "node-1"})

	parsed, err := setting.AsClusterVolume()
	assert.NoError(t, err)
	assert.False(t, parsed.Managed)
	assert.Empty(t, parsed.Driver)
	assert.Empty(t, parsed.DriverOpts)
}

func TestClusterVolumeIsPinned(t *testing.T) {
	assert.True(t, (&ClusterVolume{NodeID: "node-1"}).IsPinned())
	assert.True(t, (&ClusterVolume{NodeLabel: "storage=fast"}).IsPinned())
	// Neither is the claim that every node reaches the data, not an omission.
	assert.False(t, (&ClusterVolume{}).IsPinned())
	assert.False(t, (*ClusterVolume)(nil).IsPinned())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/entity/ -run TestClusterVolume -v`
Expected: FAIL — `unknown field Managed in struct literal`, `parsed.IsPinned undefined`.

- [ ] **Step 3: Write minimal implementation**

In `hivepaas_app/entity/setting_cluster_volume.go`, replace the `ClusterVolume` struct and add `IsPinned`:

```go
type ClusterVolume struct {
	// Which node the volume's data is on, by id or by label. Decided when the
	// volume is created and never changed afterwards: an immutable answer is
	// what lets a mount spec and a placement constraint derived from it stay
	// correct for as long as the volume exists.
	NodeID    string `json:"nodeId,omitempty"`
	NodeLabel string `json:"nodeLabel,omitempty"`

	// Managed says HivePaaS wrote the specification below and is responsible for
	// reproducing the volume on whatever node needs it. A volume that arrived
	// through discovery belongs to somebody else: it is mounted by name and
	// nothing about it is inferred.
	Managed    bool              `json:"managed,omitempty"`
	Driver     string            `json:"driver,omitempty"`
	DriverOpts map[string]string `json:"driverOpts,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// IsPinned reports whether the volume names a node it has to be on.
func (s *ClusterVolume) IsPinned() bool {
	if s == nil {
		return false
	}
	return s.NodeID != "" || s.NodeLabel != ""
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/entity/ -run TestClusterVolume -v`
Expected: PASS (4 tests)

- [ ] **Step 5: Verify nothing else broke and commit**

```bash
go build ./... && go test ./hivepaas_app/... 2>&1 | grep -v "^ok\|no test files" | head
git add hivepaas_app/entity/setting_cluster_volume.go hivepaas_app/entity/setting_cluster_volume_test.go
git commit -m "Carry the volume specification in its setting"
```

---

### Task 2: Make node pinning immutable

Reverts the update-side half of `deb5d974`. Pinning is decided at creation; update keeps only `inheritable` and `default`, which is what the note already in `UpdateVolume` describes.

**Files:**
- Modify: `hivepaas_app/usecase/cluster/volumeuc/volumedto/update.go`
- Modify: `hivepaas_app/usecase/cluster/volumeuc/update.go`
- Modify: `hivepaas_app/usecase/cluster/volumeuc/volumedto/create_test.go:65-121`

**Interfaces:**
- Consumes: Task 1's `entity.ClusterVolume`.
- Produces: `volumedto.UpdateVolumeReq` with no pinning fields and no `Pinning()` method.

- [ ] **Step 1: Write the failing test**

Replace everything from `func strPtr` to the end of `hivepaas_app/usecase/cluster/volumeuc/volumedto/create_test.go` with:

```go
func updateReq() *UpdateVolumeReq {
	req := NewUpdateVolumeReq()
	req.ID = "01JAB9XED0GTXBSQDFVYAJ8WA9"
	return req
}

// Pinning is settled when the volume is created. An update that could move it
// would be a promise HivePaaS cannot keep: the data does not follow the pin, so
// the only thing a change moves is where HivePaaS goes looking.
func TestUpdateVolumeCarriesNoPinning(t *testing.T) {
	assert.Empty(t, updateReq().Validate())

	req := updateReq()
	typ := reflect.TypeOf(*req)
	for _, field := range []string{"NodeID", "NodeLabel"} {
		if _, found := typ.FieldByName(field); found {
			t.Errorf("update request must not carry %s: pinning is immutable", field)
		}
	}
}
```

Add `"reflect"` to that file's stdlib import group.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/usecase/cluster/volumeuc/volumedto/ -v`
Expected: FAIL — the old update tests no longer compile (`strPtr` removed) and `TestUpdateVolumeCarriesNoPinning` reports both fields still present.

- [ ] **Step 3: Write minimal implementation**

In `volumedto/update.go`, delete the `NodeID`/`NodeLabel` fields with their doc comment, delete the `Pinning()` method, and reduce `Validate` to:

```go
// Validate implements interface basedto.ReqValidator
func (req *UpdateVolumeReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}
```

Drop the now-unused `basedto`, `entity` and `gofn` imports, keeping `basedto` only if `UpdateVolumeResp` still refers to it (it does — `basedto.Meta`).

In `volumeuc/update.go`, delete the `pinning` block before the transaction and the whole `PrepareUpdate` closure, leaving:

```go
func (uc *UC) UpdateVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.UpdateVolumeReq,
) (*volumedto.UpdateVolumeResp, error) {
	req.Type = currentSettingType
	req.Auth = auth

	// Only `inheritable` and `default` are updatable. Everything that describes
	// the volume itself - where its data is, and how to mount it - is settled at
	// creation, because the data does not move when the description does.
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &volumedto.UpdateVolumeResp{}, nil
}
```

Drop the now-unused `database`, `entity` and `settings`-adjacent imports that go cold (`database` and `entity` will; `settings` stays).

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/usecase/cluster/volumeuc/... -v && go build ./...`
Expected: PASS; build clean.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app/usecase/cluster/volumeuc/
git commit -m "Make volume node pinning immutable"
```

---

### Task 3: Record the specification at creation instead of calling docker

**Files:**
- Create: `hivepaas_app/usecase/cluster/volumeuc/volumedto/driver_opts.go`
- Create: `hivepaas_app/usecase/cluster/volumeuc/volumedto/driver_opts_test.go`
- Modify: `hivepaas_app/usecase/cluster/volumeuc/create.go`

**Interfaces:**
- Consumes: Task 1's `entity.ClusterVolume`.
- Produces: `func (req *VolumeBaseReq) BuildDriverOpts(bindDirectory string) map[string]string`; `VolumeBaseReq.ToEntity()` now returns a fully populated `*entity.ClusterVolume` including `Managed: true`, `Driver`, `DriverOpts`, `Labels`.

- [ ] **Step 1: Write the failing test**

Create `volumedto/driver_opts_test.go`:

```go
package volumedto

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/services/docker"
)

func TestBuildDriverOptsBind(t *testing.T) {
	req := &VolumeBaseReq{
		Driver: docker.VolumeDriverLocal,
		BindOptions: &VolumeBindOptionsReq{
			Propagation:  mount.PropagationRShared,
			Readonly:     true,
			ExtraOptions: "noatime",
		},
	}

	opts := req.BuildDriverOpts("/data/pg")

	assert.Equal(t, "none", opts["type"])
	assert.Equal(t, "/data/pg", opts["device"])
	assert.Equal(t, "bind,ro,rshared,noatime", opts["o"])
}

func TestBuildDriverOptsNfs(t *testing.T) {
	req := &VolumeBaseReq{
		Driver: docker.VolumeDriverLocal,
		NfsOptions: &VolumeNfsOptionsReq{
			Addr:    "10.0.0.5",
			Device:  ":/exports/data",
			Version: "4.1",
		},
	}

	opts := req.BuildDriverOpts("")

	assert.Equal(t, "nfs", opts["type"])
	assert.Equal(t, ":/exports/data", opts["device"])
	assert.Equal(t, "addr=10.0.0.5,rw,nfsvers=4.1", opts["o"])
}

func TestBuildDriverOptsTmpfs(t *testing.T) {
	req := &VolumeBaseReq{
		Driver:       docker.VolumeDriverLocal,
		TmpfsOptions: &VolumeTmpfsOptionsReq{Size: unit.DataSize(100 * unit.MB), UID: 1000},
	}

	opts := req.BuildDriverOpts("")

	assert.Equal(t, "tmpfs", opts["type"])
	assert.Equal(t, "size=100m,uid=1000", opts["o"])
}

// Client-supplied options may add keys but never repoint the volume: type and
// device are what decide where the data is, and the request already said that
// through the typed fields.
func TestBuildDriverOptsRefusesToRepoint(t *testing.T) {
	req := &VolumeBaseReq{
		Driver:      docker.VolumeDriverLocal,
		BindOptions: &VolumeBindOptionsReq{},
		Options:     map[string]string{"type": "nfs", "device": "/elsewhere", "extra": "kept"},
	}

	opts := req.BuildDriverOpts("/data/pg")

	assert.Equal(t, "none", opts["type"])
	assert.Equal(t, "/data/pg", opts["device"])
	assert.Equal(t, "kept", opts["extra"])
}

// A custom driver is handed its options untouched: HivePaaS knows nothing about
// what they mean.
func TestBuildDriverOptsCustomDriver(t *testing.T) {
	req := &VolumeBaseReq{
		Driver:  "some-plugin",
		Options: map[string]string{"size": "10G"},
	}

	opts := req.BuildDriverOpts("")

	assert.Equal(t, map[string]string{"size": "10G"}, opts)
}

// The entity written to the setting is the whole description, because it is the
// only copy that reaches another node.
func TestToEntityRecordsTheSpecification(t *testing.T) {
	req := &VolumeBaseReq{
		Name:        "pgdata",
		Driver:      docker.VolumeDriverLocal,
		NodeID:      "node-1",
		BindOptions: &VolumeBindOptionsReq{},
		Labels:      map[string]string{"team": "core"},
	}

	vol := req.ToEntityWithDriverOpts(req.BuildDriverOpts("/data/pg"))

	assert.Equal(t, "node-1", vol.NodeID)
	assert.True(t, vol.Managed)
	assert.Equal(t, "local", vol.Driver)
	assert.Equal(t, "/data/pg", vol.DriverOpts["device"])
	assert.Equal(t, "core", vol.Labels["team"])
	// The docker name is a ULID, so the chosen name has to travel as a label.
	assert.Equal(t, "pgdata", vol.Labels[docker.VolumeNameLabel])
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/usecase/cluster/volumeuc/volumedto/ -run "DriverOpts|ToEntity" -v`
Expected: FAIL — `req.BuildDriverOpts undefined`, `req.ToEntityWithDriverOpts undefined`.

- [ ] **Step 3: Write minimal implementation**

Create `volumedto/driver_opts.go`:

```go
package volumedto

import (
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/services/docker"
)

// BuildDriverOpts assembles the docker driver options describing this volume.
//
// bindDirectory is passed in rather than resolved here because resolving it
// touches the host filesystem, and this has to stay a function a test can call.
func (req *VolumeBaseReq) BuildDriverOpts(bindDirectory string) map[string]string {
	driverOpts := map[string]string{}

	if req.Driver == docker.VolumeDriverLocal {
		switch {
		case req.BindOptions != nil:
			driverOpts["type"] = "none"
			driverOpts["device"] = bindDirectory
			o := fmt.Sprintf("bind,%s", gofn.If(req.BindOptions.Readonly, "ro", "rw"))
			if req.BindOptions.Propagation != "" {
				o += "," + string(req.BindOptions.Propagation)
			}
			if req.BindOptions.ExtraOptions != "" {
				o += "," + req.BindOptions.ExtraOptions
			}
			driverOpts["o"] = o

		case req.NfsOptions != nil:
			driverOpts["type"] = "nfs"
			driverOpts["device"] = req.NfsOptions.Device
			o := fmt.Sprintf("addr=%s,%s", req.NfsOptions.Addr,
				gofn.If(req.NfsOptions.Readonly, "ro", "rw"))
			if req.NfsOptions.Version != "" {
				o += ",nfsvers=" + req.NfsOptions.Version
			}
			if req.NfsOptions.ExtraOptions != "" {
				o += "," + req.NfsOptions.ExtraOptions
			}
			driverOpts["o"] = o

		case req.TmpfsOptions != nil:
			driverOpts["type"] = "tmpfs"
			driverOpts["device"] = gofn.Coalesce(req.TmpfsOptions.Device, "tmpfs")
			bytes := req.TmpfsOptions.Size.Bytes() + int64(unit.MB) - 1
			o := fmt.Sprintf("size=%vm", bytes/int64(unit.MB))
			if req.TmpfsOptions.Mode > 0 {
				o += fmt.Sprintf(",mode=%v", req.TmpfsOptions.Mode)
			}
			if req.TmpfsOptions.UID > 0 {
				o += fmt.Sprintf(",uid=%v", req.TmpfsOptions.UID)
			}
			if req.TmpfsOptions.GID > 0 {
				o += fmt.Sprintf(",gid=%v", req.TmpfsOptions.GID)
			}
			driverOpts["o"] = o
		}
	}

	// Extra options from the client may add keys, but type and device are what
	// decide where the data is and the typed fields already answered that.
	for k, v := range req.Options {
		if _, ok := driverOpts[k]; ok && (k == "type" || k == "device") {
			continue
		}
		driverOpts[k] = v
	}
	return driverOpts
}

// ToEntityWithDriverOpts records the whole description of the volume, which is
// the only copy that reaches a node other than this one.
func (req *VolumeBaseReq) ToEntityWithDriverOpts(driverOpts map[string]string) *entity.ClusterVolume {
	// The volume's name in docker is a ULID, so the name a person chose has to
	// travel as a label or an operator reading `docker volume ls` on a node sees
	// nothing they recognise.
	labels := map[string]string{docker.VolumeNameLabel: req.Name}
	for k, v := range req.Labels {
		labels[k] = v
	}

	return &entity.ClusterVolume{
		NodeID:     req.NodeID,
		NodeLabel:  req.NodeLabel,
		Managed:    true,
		Driver:     string(req.Driver),
		DriverOpts: driverOpts,
		Labels:     labels,
	}
}
```

Add the label name to `services/docker/volume.go`, beside `VolumeDriverLocal`, so
both the usecase and the service layer can use it without either importing the
other:

```go
// VolumeNameLabel carries the name a person gave the volume, because its docker
// name is a ULID.
const VolumeNameLabel = "hivepaas.volume.name"
```

Note the tmpfs branch keeps `device` defaulting to `"tmpfs"`, matching `createVolumeInDocker` today.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/usecase/cluster/volumeuc/volumedto/ -v`
Expected: PASS

- [ ] **Step 5: Rewrite the creation path to record rather than create**

In `volumeuc/create.go`, replace `createVolumeInDocker` with `prepareVolumeSpec` and rewire `CreateVolume`:

```go
func (uc *UC) CreateVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.CreateVolumeReq,
) (*volumedto.CreateVolumeResp, error) {
	req.Type = currentSettingType
	req.Auth = auth

	nodeID, err := uc.resolveCurrentNode(ctx, req.NodeID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	req.NodeID = nodeID

	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		// The settings framework refuses a duplicate name for us. It replaces the
		// VolumeInspect this used to do, which asked the wrong daemon anyway.
		VerifyingName:   req.Name,
		VerifyingRefIDs: (&entity.ClusterVolume{}).GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			if req.Scope.IsProjectScope() {
				req.Name = req.Scope.Project.Key + "_" + req.Name
			}
			volEntity, err := uc.prepareVolumeSpec(ctx, req)
			if err != nil {
				return hperrors.Wrap(err)
			}
			// The setting's own id is the volume's name in docker. Nothing has
			// created it yet - docker does that when a task first mounts it - so
			// the name has to come from the only identity that exists at this
			// point.
			pData.Setting.RefID = pData.Setting.ID
			pData.Setting.Name = req.Name
			pData.Setting.Kind = volEntity.Driver
			return hperrors.Wrap(pData.Setting.SetData(volEntity))
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &volumedto.CreateVolumeResp{
		Data: resp.Data,
	}, nil
}

// prepareVolumeSpec works out the description that will be stored, doing the two
// things that have to happen on a real host: making the bind directory on the
// pinned node, and resolving the directory a bind volume points at.
func (uc *UC) prepareVolumeSpec(
	ctx context.Context,
	req *volumedto.CreateVolumeReq,
) (*entity.ClusterVolume, error) {
	isPinnedToNode := req.NodeID != "" || req.NodeLabel != ""

	if isPinnedToNode && req.BindOptions != nil && req.BindOptions.Directory != "" {
		if err := uc.createBindDirectoryInNode(ctx, req); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	bindDirectory := ""
	if req.Driver == docker.VolumeDriverLocal && req.BindOptions != nil {
		directory, err := uc.calcBindDirectory(ctx, req, req.BindOptions.Directory)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		bindDirectory = directory
	}

	return req.ToEntityWithDriverOpts(req.BuildDriverOpts(bindDirectory)), nil
}
```

Delete the old `VolumeInspect` duplicate check, the `driverOpts` assembly (now in `BuildDriverOpts`), and the `VolumeCreate` call. Delete the now-unused `ToEntity` method in `volumedto/create.go`. Remove imports that go cold in `create.go`: `errors`, `fmt`, `client`, `gofn`, `unit`, `dockerhelper`, `base`.

- [ ] **Step 6: Verify and commit**

Run: `go build ./... && go test ./hivepaas_app/usecase/cluster/volumeuc/... -v && make lint-local`
Expected: PASS, no lint findings.

```bash
git add hivepaas_app/usecase/cluster/volumeuc/
git commit -m "Record the volume specification instead of creating it in docker"
```

---

### Task 4: Same for the project default volume

**Files:**
- Modify: `hivepaas_app/service/volumeservice/volumeserviceimpl/project.go:29-90`
- Modify: `hivepaas_app/service/volumeservice/service.go:24-25`
- Modify: callers of `CreateProjectDefaultVolume` (find with the grep in Step 1)

**Interfaces:**
- Consumes: Task 1's `entity.ClusterVolume`.
- Produces: `CreateProjectDefaultVolume(ctx context.Context, project *entity.Project) (*entity.Setting, error)` — the `*client.VolumeCreateResult` return is gone.

- [ ] **Step 1: Find the callers**

Run: `grep -rn "CreateProjectDefaultVolume" hivepaas_app --include='*.go'`
Note every call site; each must drop the second return value.

- [ ] **Step 2: Write the failing test**

Create `hivepaas_app/service/volumeservice/volumeserviceimpl/project_test.go`:

```go
package volumeserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/docker"
)

// The project's default volume is described in its setting like any other, so
// the node that ends up running a task can rebuild it.
func TestProjectDefaultVolumeRecordsItsSpecification(t *testing.T) {
	setting := buildProjectDefaultVolumeSetting(
		&entity.Project{ID: "01JPROJECT0000000000000000", Key: "shop"},
		"/srv/hivepaas",
		"node-1",
	)

	assert.Equal(t, "default", setting.Name)
	assert.Equal(t, setting.ID, setting.RefID, "the setting id is the volume name")
	assert.Equal(t, "local", setting.Kind)

	vol, err := setting.AsClusterVolume()
	assert.NoError(t, err)
	assert.True(t, vol.Managed)
	assert.Equal(t, "node-1", vol.NodeID)
	assert.Equal(t, "none", vol.DriverOpts["type"])
	assert.Equal(t, "/srv/hivepaas/project_data/shop", vol.DriverOpts["device"])
	assert.Equal(t, "bind,rw", vol.DriverOpts["o"])
	assert.Equal(t, "default", vol.Labels[docker.VolumeNameLabel])
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./hivepaas_app/service/volumeservice/volumeserviceimpl/ -run TestProjectDefault -v`
Expected: FAIL — `buildProjectDefaultVolumeSetting undefined`.

- [ ] **Step 4: Write minimal implementation**

In `project.go`, extract the pure part and drop the docker call:

```go
func (s *service) CreateProjectDefaultVolume(
	ctx context.Context,
	project *entity.Project,
) (_ *entity.Setting, err error) {
	storagePathInHost := config.Current().Storage.BindSource
	if storagePathInHost == "" {
		return nil, hperrors.Wrap(hperrors.ErrUnconfigured).
			WithParam("Name", "HP_STORAGE_BIND_SOURCE")
	}

	subpath := filepath.Join("project_data", project.Key)
	err = s.MakeSubDirInHost(ctx, storagePathInHost, subpath, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	nodeID, err := s.dockerManager.NodeCurrentID(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return buildProjectDefaultVolumeSetting(project, storagePathInHost, nodeID), nil
}

// buildProjectDefaultVolumeSetting is the whole record, and nothing in it needs
// a host to produce - which is what makes it testable and what lets the volume
// be rebuilt on a node this process never talks to.
func buildProjectDefaultVolumeSetting(
	project *entity.Project,
	storagePathInHost string,
	nodeID string,
) *entity.Setting {
	timeNow := timeutil.NowUTC()
	settingID := gofn.Must(ulid.NewStringULID())

	setting := &entity.Setting{
		ID:              settingID,
		Scope:           base.ObjectScopeProject,
		ObjectID:        project.ID,
		Type:            base.SettingTypeClusterVolume,
		Kind:            string(docker.VolumeDriverLocal),
		Status:          base.SettingStatusActive,
		Name:            "default",
		RefID:           settingID,
		Inheritable:     true,
		Default:         true,
		Version:         entity.CurrentClusterVolumeVersion,
		UpdateVer:       1,
		CreatedAt:       timeNow,
		UpdatedAt:       timeNow,
		CurrentObjectID: project.ID,
	}
	setting.MustSetData(&entity.ClusterVolume{
		NodeID:  nodeID,
		Managed: true,
		Driver:  string(docker.VolumeDriverLocal),
		DriverOpts: map[string]string{
			"type":   "none",
			"device": filepath.Join(storagePathInHost, filepath.Join("project_data", project.Key)),
			"o":      "bind,rw",
		},
		Labels: map[string]string{
			docker.StackLabelNamespace: project.Key,
			docker.VolumeNameLabel:     "default",
		},
	})
	return setting
}
```

Update the interface in `volumeservice/service.go` to the two-value signature, and fix each call site found in Step 1 to drop the `*client.VolumeCreateResult`. Remove the `client` import from `project.go` if it goes cold.

- [ ] **Step 5: Run test to verify it passes**

Run: `go build ./... && go test ./hivepaas_app/service/volumeservice/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add hivepaas_app/service/volumeservice/ hivepaas_app/service/projectservice/
git commit -m "Record the project default volume without creating it in docker"
```

---

### Task 5: Build mount specs from the setting

**Files:**
- Modify: `hivepaas_app/usecase/appsettingsuc/storage_settings_update.go:136-258`
- Create: `hivepaas_app/usecase/appsettingsuc/volume_mount.go`
- Create: `hivepaas_app/usecase/appsettingsuc/volume_mount_test.go`

**Interfaces:**
- Consumes: Task 1's `entity.ClusterVolume`.
- Produces: `func applyVolumeDriverConfig(dockerMnt *mount.Mount, vol *entity.ClusterVolume)` and `func bindMountTarget(vol *entity.ClusterVolume, subpath string) (directory string, propagation mount.Propagation, ok bool)`.

- [ ] **Step 1: Write the failing test**

Create `hivepaas_app/usecase/appsettingsuc/volume_mount_test.go`:

```go
package appsettingsuc

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A bind volume is mounted by path, so the mount carries everything the node
// needs without consulting any daemon.
func TestBindMountTargetFromSetting(t *testing.T) {
	vol := &entity.ClusterVolume{
		Managed: true,
		Driver:  "local",
		DriverOpts: map[string]string{
			"type": "none", "device": "/srv/data", "o": "bind,rw,rshared",
		},
	}

	directory, propagation, ok := bindMountTarget(vol, "shop/prod/web")

	assert.True(t, ok)
	assert.Equal(t, "/srv/data/shop/prod/web", directory)
	assert.Equal(t, mount.PropagationRShared, propagation)
}

func TestBindMountTargetRejectsNonBindVolumes(t *testing.T) {
	tests := []struct {
		name string
		vol  *entity.ClusterVolume
	}{
		{"nfs", &entity.ClusterVolume{Driver: "local", DriverOpts: map[string]string{"type": "nfs", "device": ":/e"}}},
		{"no device", &entity.ClusterVolume{Driver: "local", DriverOpts: map[string]string{"type": "none"}}},
		{"custom driver", &entity.ClusterVolume{Driver: "some-plugin"}},
		{"discovered, nothing recorded", &entity.ClusterVolume{}},
		{"nil", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, ok := bindMountTarget(tt.vol, "shop/prod/web")
			assert.False(t, ok)
		})
	}
}

// Driver config in the mount spec is what makes a node materialize the volume
// correctly instead of inventing an empty one.
func TestApplyVolumeDriverConfig(t *testing.T) {
	dockerMnt := &mount.Mount{Type: mount.TypeVolume, Source: "01JVOL"}
	vol := &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "nfs", "device": ":/exports/data", "o": "addr=10.0.0.5,rw"},
		Labels:     map[string]string{"hivepaas.volume.name": "shared"},
	}

	applyVolumeDriverConfig(dockerMnt, vol)

	assert.NotNil(t, dockerMnt.VolumeOptions)
	assert.Equal(t, "local", dockerMnt.VolumeOptions.DriverConfig.Name)
	assert.Equal(t, ":/exports/data", dockerMnt.VolumeOptions.DriverConfig.Options["device"])
	assert.Equal(t, "shared", dockerMnt.VolumeOptions.Labels["hivepaas.volume.name"])
}

// A volume HivePaaS did not author is mounted by name and nothing is inferred
// about it: stamping a specification we did not write would be a guess.
func TestApplyVolumeDriverConfigSkipsUnmanaged(t *testing.T) {
	dockerMnt := &mount.Mount{Type: mount.TypeVolume, Source: "someones-volume"}
	vol := &entity.ClusterVolume{
		Managed:    false,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "nfs"},
	}

	applyVolumeDriverConfig(dockerMnt, vol)

	assert.Nil(t, dockerMnt.VolumeOptions)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/usecase/appsettingsuc/ -run "BindMountTarget|ApplyVolumeDriverConfig" -v`
Expected: FAIL — `bindMountTarget undefined`, `applyVolumeDriverConfig undefined`.

- [ ] **Step 3: Write minimal implementation**

Create `hivepaas_app/usecase/appsettingsuc/volume_mount.go`:

```go
package appsettingsuc

import (
	"path/filepath"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/docker"
)

// bindMountTarget reports the directory a bind volume points at, so it can be
// mounted by path rather than by name.
//
// It reads the setting rather than the docker volume because the docker volume
// only exists on the node that happened to create it, which is the whole reason
// a volume's description lives in its setting now.
func bindMountTarget(
	vol *entity.ClusterVolume,
	subpath string,
) (directory string, propagation mount.Propagation, ok bool) {
	if vol == nil || vol.Driver != string(docker.VolumeDriverLocal) {
		return "", "", false
	}
	device := vol.DriverOpts["device"]
	if vol.DriverOpts["type"] != "none" || device == "" {
		return "", "", false
	}
	return filepath.Join(device, subpath), getConfiguredPropagation(vol.DriverOpts["o"]), true
}

// applyVolumeDriverConfig puts the volume's description into the mount, so the
// node running the task builds the volume from it instead of creating an empty
// default when the name is unknown there.
func applyVolumeDriverConfig(dockerMnt *mount.Mount, vol *entity.ClusterVolume) {
	if vol == nil || !vol.Managed || vol.Driver == "" {
		return
	}
	if dockerMnt.VolumeOptions == nil {
		dockerMnt.VolumeOptions = &mount.VolumeOptions{}
	}
	dockerMnt.VolumeOptions.DriverConfig = &mount.Driver{
		Name:    vol.Driver,
		Options: vol.DriverOpts,
	}
	if len(vol.Labels) > 0 && dockerMnt.VolumeOptions.Labels == nil {
		dockerMnt.VolumeOptions.Labels = vol.Labels
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/usecase/appsettingsuc/ -run "BindMountTarget|ApplyVolumeDriverConfig" -v`
Expected: PASS (7 subtests)

- [ ] **Step 5: Rewire `storage_settings_update.go` to use them**

Delete the `VolumeListByIDs` call and the `DockerVolumes` map from `loadUpdatingAppStorageSettingsData` and from the `updateAppStorageSettingsData` struct. In `prepareUpdatingAppStorageSettings`, replace the `dockerVol` lookup:

```go
func (uc *UC) prepareUpdatingAppStorageSettings(
	ctx context.Context,
	data *updateAppStorageSettingsData,
) error {
	for _, reqMnt := range data.NewMountReqs {
		dbVol := data.DBVolumes[reqMnt.Source]
		vol, err := dbVol.AsClusterVolume()
		if err != nil {
			return hperrors.Wrap(err)
		}
		dockerMnt := &mount.Mount{
			Type:        reqMnt.Type,
			Source:      dbVol.RefID,
			Target:      reqMnt.Target,
			ReadOnly:    reqMnt.ReadOnly,
			Consistency: reqMnt.Consistency,
		}

		uc.buildDockerMount(ctx, dockerMnt, reqMnt, vol, dbVol, data)

		if dockerMnt.Type == mount.TypeVolume || dockerMnt.Type == mount.TypeCluster {
			subpath := ""
			if dockerMnt.VolumeOptions != nil {
				subpath = dockerMnt.VolumeOptions.Subpath
			}
			_ = uc.volumeService.EnsureVolumePermissions(ctx, dockerMnt, subpath)
		}

		data.FinalMounts = append(data.FinalMounts, *dockerMnt)
	}
	return nil
}
```

Change `buildDockerMount`'s `dockerVol *volume.Volume` parameter to `vol *entity.ClusterVolume`, replace the `useBindMountIfAppropriate(ctx, dockerMnt, dockerVol, subpath)` call, and add the driver config in the `TypeVolume` branch:

```go
	case mount.TypeVolume:
		if reqMnt.VolumeOptions != nil {
			dockerMnt.VolumeOptions = &mount.VolumeOptions{
				Subpath: subpath,
				NoCopy:  reqMnt.VolumeOptions.NoCopy,
				Labels:  reqMnt.VolumeOptions.Labels,
			}
			if reqMnt.VolumeOptions.DriverConfig != nil {
				dockerMnt.VolumeOptions.DriverConfig = &mount.Driver{
					Name:    reqMnt.VolumeOptions.DriverConfig.Name,
					Options: reqMnt.VolumeOptions.DriverConfig.Options,
				}
			}
		}
		// Only when the request did not bring its own: an explicit DriverConfig
		// from the client is the caller overriding, and this must not undo it.
		if dockerMnt.VolumeOptions == nil || dockerMnt.VolumeOptions.DriverConfig == nil {
			applyVolumeDriverConfig(dockerMnt, vol)
		}
```

Rewrite `useBindMountIfAppropriate` to take the entity:

```go
func (uc *UC) useBindMountIfAppropriate(
	ctx context.Context,
	dockerMnt *mount.Mount,
	vol *entity.ClusterVolume,
	subpath string,
) {
	directory, propagation, ok := bindMountTarget(vol, subpath)
	if !ok {
		return
	}
	if err := uc.volumeService.MakeSubDirInHost(ctx, vol.DriverOpts["device"], subpath, true); err != nil {
		return
	}

	dockerMnt.Type = mount.TypeBind
	dockerMnt.Source = directory
	dockerMnt.BindOptions = &mount.BindOptions{
		// Kept for volumes with no pin, which claim every node reaches the same
		// data: creating the directory there is the right thing. A pinned volume
		// is kept on its node by a placement constraint instead.
		CreateMountpoint: true,
	}
	if propagation != "" {
		dockerMnt.BindOptions.Propagation = propagation
	}
	dockerMnt.VolumeOptions = nil
	dockerMnt.ClusterOptions = nil
	dockerMnt.TmpfsOptions = nil
	dockerMnt.ImageOptions = nil
}
```

Make `prepareUpdatingAppStorageSettings`'s new `error` return propagate at its call site.

- [ ] **Step 6: Verify and commit**

Run: `go build ./... && go test ./hivepaas_app/usecase/appsettingsuc/... -v && make lint-local`
Expected: PASS

```bash
git add hivepaas_app/usecase/appsettingsuc/
git commit -m "Build volume mounts from the setting rather than from docker"
```

---

### Task 6: Turn a pin into a placement constraint

**Files:**
- Create: `hivepaas_app/service/placementservice/volume_pin.go`
- Create: `hivepaas_app/service/placementservice/volume_pin_test.go`
- Modify: `hivepaas_app/service/placementservice/types.go`
- Modify: `hivepaas_app/service/placementservice/placementserviceimpl/placement.go`
- Modify: `hivepaas_app/service/placementservice/placementserviceimpl/placement_apply.go:88`
- Create: `hivepaas_app/service/placementservice/placementserviceimpl/placement_apply_test.go`

**Interfaces:**
- Consumes: Task 1's `entity.ClusterVolume`.
- Produces: `placementservice.VolumePin{VolumeName, NodeID, NodeLabel string}`; `func placementservice.VolumePinConstraint(pins []VolumePin) (string, *VolumePinConflict)`; `VolumePinConflict{First, Second VolumePin}` with an `Error() string`; `ApplyPlacementSettingsReq.VolumePins []VolumePin`.

- [ ] **Step 1: Write the failing test**

Create `hivepaas_app/service/placementservice/volume_pin_test.go`:

```go
package placementservice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVolumePinConstraintByID(t *testing.T) {
	constraint, conflict := VolumePinConstraint([]VolumePin{{VolumeName: "pgdata", NodeID: "node-1"}})

	assert.Nil(t, conflict)
	assert.Equal(t, "node.id==node-1", constraint)
}

func TestVolumePinConstraintByLabel(t *testing.T) {
	constraint, conflict := VolumePinConstraint([]VolumePin{
		{VolumeName: "pgdata", NodeLabel: "storage=fast"},
	})

	assert.Nil(t, conflict)
	assert.Equal(t, "node.labels.storage==fast", constraint)
}

// A label with no value means the key has to be present, which swarm spells as
// the string "true".
func TestVolumePinConstraintBareLabel(t *testing.T) {
	constraint, conflict := VolumePinConstraint([]VolumePin{{VolumeName: "d", NodeLabel: "ssd"}})

	assert.Nil(t, conflict)
	assert.Equal(t, "node.labels.ssd==true", constraint)
}

// Volumes with no pin say every node reaches their data, so they constrain
// nothing.
func TestVolumePinConstraintIgnoresUnpinned(t *testing.T) {
	constraint, conflict := VolumePinConstraint([]VolumePin{
		{VolumeName: "shared"},
		{VolumeName: "pgdata", NodeID: "node-1"},
	})

	assert.Nil(t, conflict)
	assert.Equal(t, "node.id==node-1", constraint)
}

func TestVolumePinConstraintAgreeingPins(t *testing.T) {
	constraint, conflict := VolumePinConstraint([]VolumePin{
		{VolumeName: "pgdata", NodeID: "node-1"},
		{VolumeName: "uploads", NodeID: "node-1"},
	})

	assert.Nil(t, conflict)
	assert.Equal(t, "node.id==node-1", constraint)
}

// Two volumes on two different nodes cannot both be reached by one task. This is
// a real contradiction, so it is reported rather than handed to swarm as a set
// of constraints no node satisfies.
func TestVolumePinConstraintConflict(t *testing.T) {
	_, conflict := VolumePinConstraint([]VolumePin{
		{VolumeName: "pgdata", NodeID: "node-1"},
		{VolumeName: "uploads", NodeID: "node-2"},
	})

	assert.NotNil(t, conflict)
	assert.Contains(t, conflict.Error(), "pgdata")
	assert.Contains(t, conflict.Error(), "uploads")
}

func TestVolumePinConstraintConflictAcrossIDAndLabel(t *testing.T) {
	_, conflict := VolumePinConstraint([]VolumePin{
		{VolumeName: "pgdata", NodeID: "node-1"},
		{VolumeName: "uploads", NodeLabel: "storage=fast"},
	})

	assert.NotNil(t, conflict)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/service/placementservice/ -v`
Expected: FAIL — `VolumePin undefined`, `VolumePinConstraint undefined`.

- [ ] **Step 3: Write minimal implementation**

Create `hivepaas_app/service/placementservice/volume_pin.go`:

```go
package placementservice

import (
	"fmt"
	"strings"
)

// VolumePin is a volume's claim about which node its data is on.
type VolumePin struct {
	VolumeName string
	NodeID     string
	NodeLabel  string
}

func (p VolumePin) isPinned() bool {
	return p.NodeID != "" || p.NodeLabel != ""
}

// constraint is the swarm spelling of this pin, empty when there is none.
func (p VolumePin) constraint() string {
	if p.NodeID != "" {
		return "node.id==" + p.NodeID
	}
	if p.NodeLabel == "" {
		return ""
	}
	key, val, found := strings.Cut(p.NodeLabel, "=")
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	val = strings.TrimSpace(val)
	if !found || val == "" {
		// A bare key means "this label is set", which swarm writes as "true".
		val = "true"
	}
	return fmt.Sprintf("node.labels.%s==%s", key, val)
}

// VolumePinConflict is two volumes that cannot both be reached from one node.
type VolumePinConflict struct {
	First  VolumePin
	Second VolumePin
}

func (c *VolumePinConflict) Error() string {
	return fmt.Sprintf("volume '%s' is pinned to %s and volume '%s' to %s, so no node has both",
		c.First.VolumeName, c.First.constraint(),
		c.Second.VolumeName, c.Second.constraint())
}

// VolumePinConstraint is the single placement constraint the given volumes
// require, or a conflict when they require different ones.
//
// Unlike every other constraint HivePaaS emits this one is required rather than
// an exclusion, so two of them disagreeing is unsatisfiable rather than merely
// narrow - which is why the disagreement is returned instead of applied.
func VolumePinConstraint(pins []VolumePin) (string, *VolumePinConflict) {
	var chosen VolumePin
	for _, pin := range pins {
		if !pin.isPinned() || pin.constraint() == "" {
			continue
		}
		if !chosen.isPinned() {
			chosen = pin
			continue
		}
		if chosen.constraint() != pin.constraint() {
			return "", &VolumePinConflict{First: chosen, Second: pin}
		}
	}
	return chosen.constraint(), nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/service/placementservice/ -v`
Expected: PASS (7 tests)

- [ ] **Step 5: Write the failing test for applying it**

Create `hivepaas_app/service/placementservice/placementserviceimpl/placement_apply_test.go`:

```go
package placementserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
)

func applyData(pins []placementservice.VolumePin, currConstraints []string) *placementSettingsData {
	return &placementSettingsData{
		ApplyPlacementSettingsReq: &placementservice.ApplyPlacementSettingsReq{
			Service: &swarm.Service{
				Spec: swarm.ServiceSpec{
					TaskTemplate: swarm.TaskSpec{
						Placement: &swarm.Placement{Constraints: currConstraints},
					},
				},
			},
			PlacementSettings: &entity.AppPlacementSettings{},
			BuildSettings:     &entity.ImageBuildSettings{},
			VolumePins:        pins,
		},
	}
}

// A task can only run where its data is, so the pin becomes a constraint swarm
// enforces rather than a note HivePaaS keeps.
func TestApplyAddsVolumePinConstraint(t *testing.T) {
	data := applyData([]placementservice.VolumePin{{VolumeName: "pgdata", NodeID: "node-1"}}, nil)

	(&service{}).applyPlacementSettings(data)

	assert.Contains(t, data.Service.Spec.TaskTemplate.Placement.Constraints, "node.id==node-1")
	assert.Contains(t,
		data.Service.Spec.Labels["hivepaas.app.placementConstraints"], "node.id==node-1")
	assert.True(t, data.HasChanges)
}

// Constraints the operator set by hand survive, the same way they already do for
// the exclusions HivePaaS manages.
func TestApplyKeepsOperatorConstraints(t *testing.T) {
	data := applyData(
		[]placementservice.VolumePin{{VolumeName: "pgdata", NodeID: "node-1"}},
		[]string{"node.labels.tier==gold"},
	)

	(&service{}).applyPlacementSettings(data)

	assert.Contains(t, data.Service.Spec.TaskTemplate.Placement.Constraints, "node.labels.tier==gold")
	assert.Contains(t, data.Service.Spec.TaskTemplate.Placement.Constraints, "node.id==node-1")
}

func TestApplyAddsNothingForUnpinnedVolumes(t *testing.T) {
	data := applyData([]placementservice.VolumePin{{VolumeName: "shared"}}, nil)

	(&service{}).applyPlacementSettings(data)

	assert.Empty(t, data.Service.Spec.TaskTemplate.Placement.Constraints)
}
```

- [ ] **Step 6: Run test to verify it fails**

Run: `go test ./hivepaas_app/service/placementservice/placementserviceimpl/ -v`
Expected: FAIL — `VolumePins` is not a field of `ApplyPlacementSettingsReq`.

- [ ] **Step 7: Wire it into the apply path**

In `placementservice/types.go`, add to `ApplyPlacementSettingsReq`:

```go
	// VolumePins are the pins of the volumes this app mounts. Nil means they are
	// loaded from the app's mounts.
	VolumePins []VolumePin
```

In `placement_apply.go`, after the `ExcludeBuildNodes` block and before `finalConstraints = append(...)`:

```go
	// The one required constraint HivePaaS emits. A conflict is refused where the
	// mounts are chosen, so by here the pins already agree.
	if constraint, conflict := placementservice.VolumePinConstraint(data.VolumePins); conflict == nil &&
		constraint != "" {
		newHivepaasConstraints = append(newHivepaasConstraints, constraint)
	}
```

Add the `placementservice` import to `placement_apply.go`.

In `placement.go`, load the pins inside `loadPlacementSettingsData` after the build settings block, so a caller that did not supply them still gets them:

```go
	// Load the pins of the volumes this app mounts
	if data.VolumePins == nil {
		pins, err := s.loadVolumePins(ctx, db, data)
		if err != nil {
			return hperrors.Wrap(err)
		}
		data.VolumePins = pins
	}

	return nil
}

// loadVolumePins reads the pin of every volume the service mounts by name.
//
// Mounts are matched on RefID because that is the volume's docker-side identity,
// which is what a mount spec names.
func (s *service) loadVolumePins(
	ctx context.Context,
	db database.IDB,
	data *placementSettingsData,
) ([]placementservice.VolumePin, error) {
	var refIDs []string
	for _, mnt := range data.Service.Spec.TaskTemplate.ContainerSpec.Mounts {
		if mnt.Type == mount.TypeVolume || mnt.Type == mount.TypeCluster {
			refIDs = append(refIDs, mnt.Source)
		}
	}
	if len(refIDs) == 0 {
		return nil, nil
	}

	settings, _, err := s.settingRepo.List(ctx, db, data.App.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterVolume),
		bunex.SelectWhere("setting.ref_id IN (?)", bun.In(refIDs)),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	pins := make([]placementservice.VolumePin, 0, len(settings))
	for _, setting := range settings {
		vol, err := setting.AsClusterVolume()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		pins = append(pins, placementservice.VolumePin{
			VolumeName: setting.Name,
			NodeID:     vol.NodeID,
			NodeLabel:  vol.NodeLabel,
		})
	}
	return pins, nil
}
```

Add imports to `placement.go`: `github.com/moby/moby/api/types/mount`, `github.com/uptrace/bun`, `.../pkg/bunex`, `.../service/placementservice`. Note `loadPlacementSettingsData` returns early when `!isMultiNode` — leave that, since a single-node cluster has nowhere else to schedule and needs no constraint.

- [ ] **Step 8: Run tests to verify they pass**

Run: `go build ./... && go test ./hivepaas_app/service/placementservice/... -v`
Expected: PASS

- [ ] **Step 9: Commit**

```bash
git add hivepaas_app/service/placementservice/
git commit -m "Constrain placement to the node a mounted volume is pinned to"
```

---

### Task 7: Refuse contradictory pins when storage settings are saved

**Files:**
- Modify: `hivepaas_app/usecase/appsettingsuc/storage_settings_update.go` (the validation block ending at line ~146)
- Modify: `hivepaas_app/usecase/appsettingsuc/volume_mount_test.go`

**Interfaces:**
- Consumes: Task 6's `placementservice.VolumePinConstraint`, `VolumePin`, `VolumePinConflict`; Task 1's `entity.ClusterVolume`.
- Produces: `func volumePinsFromSettings(settings map[string]*entity.Setting) ([]placementservice.VolumePin, error)`.

- [ ] **Step 1: Write the failing test**

Append to `hivepaas_app/usecase/appsettingsuc/volume_mount_test.go`:

```go
func volumeSetting(name string, vol *entity.ClusterVolume) *entity.Setting {
	setting := &entity.Setting{
		ID:    "01J" + name,
		Name:  name,
		RefID: "01J" + name,
		Type:  base.SettingTypeClusterVolume,
	}
	setting.MustSetData(vol)
	return setting
}

func TestVolumePinsFromSettings(t *testing.T) {
	pins, err := volumePinsFromSettings(map[string]*entity.Setting{
		"a": volumeSetting("pgdata", &entity.ClusterVolume{NodeID: "node-1"}),
		"b": volumeSetting("shared", &entity.ClusterVolume{}),
	})

	assert.NoError(t, err)
	assert.Len(t, pins, 2)

	constraint, conflict := placementservice.VolumePinConstraint(pins)
	assert.Nil(t, conflict)
	assert.Equal(t, "node.id==node-1", constraint)
}

// An app cannot mount two volumes whose data is on two different nodes. Saying
// so when the mounts are chosen beats leaving swarm with a task that can never
// be scheduled and no explanation.
func TestVolumePinsFromSettingsSurfacesConflict(t *testing.T) {
	pins, err := volumePinsFromSettings(map[string]*entity.Setting{
		"a": volumeSetting("pgdata", &entity.ClusterVolume{NodeID: "node-1"}),
		"b": volumeSetting("uploads", &entity.ClusterVolume{NodeID: "node-2"}),
	})
	assert.NoError(t, err)

	_, conflict := placementservice.VolumePinConstraint(pins)
	assert.NotNil(t, conflict)
}
```

Add `base` and `placementservice` to the imports of that test file.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/usecase/appsettingsuc/ -run VolumePinsFromSettings -v`
Expected: FAIL — `volumePinsFromSettings undefined`.

- [ ] **Step 3: Write minimal implementation**

Append to `hivepaas_app/usecase/appsettingsuc/volume_mount.go`:

```go
// volumePinsFromSettings reads the pin out of each volume setting.
func volumePinsFromSettings(
	settings map[string]*entity.Setting,
) ([]placementservice.VolumePin, error) {
	pins := make([]placementservice.VolumePin, 0, len(settings))
	for _, setting := range settings {
		vol, err := setting.AsClusterVolume()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		pins = append(pins, placementservice.VolumePin{
			VolumeName: setting.Name,
			NodeID:     vol.NodeID,
			NodeLabel:  vol.NodeLabel,
		})
	}
	return pins, nil
}
```

Add `hperrors` and `placementservice` imports.

In `storage_settings_update.go`, at the end of `loadUpdatingAppStorageSettingsData` (after `data.DBVolumes = dbVolMap`), refuse the contradiction:

```go
	pins, err := volumePinsFromSettings(dbVolMap)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if _, conflict := placementservice.VolumePinConstraint(pins); conflict != nil {
		return hperrors.NewArgumentInvalid("Mounts").WithMsgLog("%s", conflict.Error())
	}

	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go build ./... && go test ./hivepaas_app/usecase/appsettingsuc/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app/usecase/appsettingsuc/
git commit -m "Refuse mounting volumes pinned to different nodes"
```

---

### Task 8: Teach sync that "not materialized" is not "deleted"

**Files:**
- Modify: `hivepaas_app/service/volumeservice/volumeserviceimpl/volumes_sync.go`
- Modify: `hivepaas_app/service/volumeservice/volumeserviceimpl/volumes_sync_test.go`

**Interfaces:**
- Consumes: Task 1's `entity.ClusterVolume`.
- Produces: sync no longer marks a setting deleted merely for being absent from docker; it backfills `Driver`/`DriverOpts`/`Labels`/`Managed` for settings that have none.

- [ ] **Step 1: Write the failing test**

In `volumes_sync_test.go`, replace `TestSyncVolumesMarksVanishedVolumesDeleted` with these, and extend `fakeDockerManager` so listed volumes can carry options:

```go
func localVolumeWithOpts(name string, opts map[string]string) volume.Volume {
	return volume.Volume{Name: name, Driver: "local", Options: opts}
}

// A volume nobody has mounted yet does not exist in docker, and that is the
// normal state of a freshly created one - not evidence that it was removed.
func TestSyncVolumesKeepsUnmaterializedSettings(t *testing.T) {
	stored := storedVolume("pgdata", &entity.ClusterVolume{
		NodeID: "node-1", Managed: true, Driver: "local",
		DriverOpts: map[string]string{"type": "none", "device": "/srv/pgdata"},
	})

	uc, repo := newSyncTest(nil, []*entity.Setting{stored})

	_, err := uc.SyncVolumes(context.Background(), nil)
	assert.NoError(t, err)

	assert.Empty(t, repo.upserted, "a volume that was never materialized was not deleted")
	assert.True(t, stored.DeletedAt.IsZero())
}

// A setting recorded before the specification was stored gets it filled in from
// the volume as it exists, so it can be rebuilt on a node this process does not
// talk to.
func TestSyncVolumesBackfillsTheSpecification(t *testing.T) {
	stored := storedVolume("pgdata", &entity.ClusterVolume{NodeID: "node-1"})
	opts := map[string]string{"type": "none", "device": "/srv/pgdata", "o": "bind,rw"}

	uc, repo := newSyncTest([]volume.Volume{localVolumeWithOpts("pgdata", opts)},
		[]*entity.Setting{stored})

	_, err := uc.SyncVolumes(context.Background(), nil)
	assert.NoError(t, err)

	setting := upsertedByName(repo, "pgdata")
	assert.NotNil(t, setting)

	parsed, err := setting.AsClusterVolume()
	assert.NoError(t, err)
	assert.True(t, parsed.Managed)
	assert.Equal(t, "local", parsed.Driver)
	assert.Equal(t, "/srv/pgdata", parsed.DriverOpts["device"])
	assert.Equal(t, "node-1", parsed.NodeID, "the operator's pin is untouched")
}

// The rule the sync already follows, now covering one more field.
func TestSyncVolumesNeverOverwritesARecordedSpecification(t *testing.T) {
	stored := storedVolume("pgdata", &entity.ClusterVolume{
		NodeID: "node-1", Managed: true, Driver: "local",
		DriverOpts: map[string]string{"type": "none", "device": "/srv/original"},
	})

	uc, _ := newSyncTest([]volume.Volume{
		localVolumeWithOpts("pgdata", map[string]string{"type": "none", "device": "/srv/moved"}),
	}, []*entity.Setting{stored})

	_, err := uc.SyncVolumes(context.Background(), nil)
	assert.NoError(t, err)

	parsed, err := stored.AsClusterVolume()
	assert.NoError(t, err)
	assert.Equal(t, "/srv/original", parsed.DriverOpts["device"])
}

// A cluster volume belongs to swarm, not to HivePaaS, so nothing is claimed for
// it.
func TestSyncVolumesDoesNotClaimClusterVolumes(t *testing.T) {
	stored := storedVolume("shared", &entity.ClusterVolume{})

	uc, _ := newSyncTest([]volume.Volume{clusterVolume("shared")}, []*entity.Setting{stored})

	_, err := uc.SyncVolumes(context.Background(), nil)
	assert.NoError(t, err)

	parsed, err := stored.AsClusterVolume()
	assert.NoError(t, err)
	assert.False(t, parsed.Managed)
	assert.Empty(t, parsed.Driver)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/service/volumeservice/volumeserviceimpl/ -run TestSyncVolumes -v`
Expected: FAIL — the unmaterialized setting is marked deleted; no backfill happens.

- [ ] **Step 3: Write minimal implementation**

In `volumes_sync.go`: delete the branch that marks a setting deleted when its volume is missing from docker, and add backfill to the branch handling settings that exist in both. Sketch of the loop over existing settings:

```go
		hasChanged := false
		if setting.Kind != vol.Driver {
			setting.Kind = vol.Driver
			hasChanged = true
		}
		if setting.Name != vol.Name {
			setting.Name = vol.Name
			hasChanged = true
		}
		changed, err := s.backfillVolumeSpec(setting, vol)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		hasChanged = hasChanged || changed
```

and the backfill itself:

```go
// backfillVolumeSpec records how a volume is built, for settings written before
// that was stored.
//
// It fills a gap and never corrects one: the pin and the specification are the
// operator's answers, and a sync that rewrote them would overwrite the answer
// every pass. A cluster volume is left alone entirely - swarm owns it, and
// HivePaaS claiming to be able to rebuild it would be a claim it cannot keep.
func (s *service) backfillVolumeSpec(setting *entity.Setting, vol *volume.Volume) (bool, error) {
	if vol.ClusterVolume != nil && vol.ClusterVolume.ID != "" {
		return false, nil
	}

	current, err := setting.AsClusterVolume()
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	if current == nil {
		current = &entity.ClusterVolume{}
	}
	if current.Driver != "" {
		return false, nil
	}

	current.Managed = true
	current.Driver = vol.Driver
	current.DriverOpts = vol.Options
	current.Labels = vol.Labels

	if err := setting.SetData(current); err != nil {
		return false, hperrors.Wrap(err)
	}
	return true, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./hivepaas_app/service/volumeservice/volumeserviceimpl/ -v`
Expected: PASS — including the pre-existing `TestSyncVolumesPinsANewlyDiscoveredVolume`, `TestSyncVolumesLeavesAClusterVolumeUnpinned`, `TestSyncVolumesNeverOverwritesRecordedPinning`, `TestSyncVolumesCarriesOverWhatDockerOwns`.

- [ ] **Step 5: Commit**

```bash
git add hivepaas_app/service/volumeservice/volumeserviceimpl/
git commit -m "Stop reading an unmaterialized volume as a deleted one"
```

---

### Task 9: Warn when the constraint would move a running task

**Files:**
- Create: `hivepaas_app/service/placementservice/placementserviceimpl/pin_drift.go`
- Create: `hivepaas_app/service/placementservice/placementserviceimpl/pin_drift_test.go`
- Modify: `hivepaas_app/service/placementservice/placementserviceimpl/placement.go`
- Modify: `hivepaas_app/service/placementservice/placementserviceimpl/service.go`

**Interfaces:**
- Consumes: Task 6's `placementservice.VolumePin` and `VolumePinConstraint`.
- Produces: `func driftingNodes(tasks []swarm.Task, pinnedNodeID string) []string`.

Constraints take effect when an app's service spec is next written; there is no sweep, because moving a task that is running today may move it away from the data it has been accumulating. The warning goes exactly where an operator is about to be affected: the moment the constraint is added.

- [ ] **Step 1: Write the failing test**

Create `pin_drift_test.go`:

```go
package placementserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
)

func runningTask(nodeID string) swarm.Task {
	return swarm.Task{NodeID: nodeID, Status: swarm.TaskStatus{State: swarm.TaskStateRunning}}
}

// Applying the pin moves this task. Saying so beats an app that quietly restarts
// somewhere else and finds a directory that is not the one it was using.
func TestDriftingNodesReportsATaskElsewhere(t *testing.T) {
	assert.Equal(t, []string{"node-2"},
		driftingNodes([]swarm.Task{runningTask("node-2")}, "node-1"))
}

func TestDriftingNodesIgnoresTasksAlreadyInPlace(t *testing.T) {
	assert.Empty(t, driftingNodes([]swarm.Task{runningTask("node-1")}, "node-1"))
}

// A task that is not running is not being moved off anything.
func TestDriftingNodesIgnoresTasksNotRunning(t *testing.T) {
	shutdown := swarm.Task{NodeID: "node-2", Status: swarm.TaskStatus{State: swarm.TaskStateShutdown}}

	assert.Empty(t, driftingNodes([]swarm.Task{shutdown}, "node-1"))
}

// Nothing to compare against: an unpinned volume constrains nothing, and a label
// pin names a set of nodes that a node id alone cannot be checked against.
func TestDriftingNodesSaysNothingWithoutANodeID(t *testing.T) {
	assert.Empty(t, driftingNodes([]swarm.Task{runningTask("node-2")}, ""))
}

func TestDriftingNodesReportsEachNodeOnce(t *testing.T) {
	tasks := []swarm.Task{runningTask("node-2"), runningTask("node-2"), runningTask("node-3")}

	assert.ElementsMatch(t, []string{"node-2", "node-3"}, driftingNodes(tasks, "node-1"))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hivepaas_app/service/placementservice/placementserviceimpl/ -run TestDriftingNodes -v`
Expected: FAIL — `driftingNodes undefined`.

- [ ] **Step 3: Write minimal implementation**

Create `pin_drift.go`:

```go
package placementserviceimpl

import (
	"github.com/moby/moby/api/types/swarm"
)

// driftingNodes are the nodes running tasks that a pin to pinnedNodeID would
// move them off.
//
// Only an id pin can be checked this way. A label pin names a set of nodes, and
// a node id on its own cannot say whether it is in that set - so an empty
// pinnedNodeID reports nothing rather than a mismatch it cannot substantiate.
func driftingNodes(tasks []swarm.Task, pinnedNodeID string) []string {
	if pinnedNodeID == "" {
		return nil
	}

	seen := map[string]struct{}{}
	var nodes []string
	for _, task := range tasks {
		if task.Status.State != swarm.TaskStateRunning || task.NodeID == "" {
			continue
		}
		if task.NodeID == pinnedNodeID {
			continue
		}
		if _, found := seen[task.NodeID]; found {
			continue
		}
		seen[task.NodeID] = struct{}{}
		nodes = append(nodes, task.NodeID)
	}
	return nodes
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hivepaas_app/service/placementservice/placementserviceimpl/ -run TestDriftingNodes -v`
Expected: PASS (5 tests)

- [ ] **Step 5: Log it where the constraint is decided**

Add a `logger logging.Logger` field and constructor parameter to `placementserviceimpl.New` and `service`, following the shape used by `settingsrevertserviceimpl` (fx wires the new parameter with no registration change).

In `placement.go`, at the end of `loadPlacementSettingsData` right after the pins are loaded:

```go
	s.warnOnPinDrift(ctx, data)
```

and in `pin_drift.go`:

```go
// warnOnPinDrift says when applying the pin will move a task that is running.
//
// It only reports. Moving the task is what the constraint does, and it happens
// when the service spec is next written - which is a moment the operator chose,
// rather than a sweep that relocates running apps on upgrade.
func (s *service) warnOnPinDrift(ctx context.Context, data *placementSettingsData) {
	constraint, conflict := placementservice.VolumePinConstraint(data.VolumePins)
	if conflict != nil || constraint == "" {
		return
	}
	var pinnedNodeID string
	for _, pin := range data.VolumePins {
		if pin.NodeID != "" {
			pinnedNodeID = pin.NodeID
			break
		}
	}
	if pinnedNodeID == "" || data.Service.ID == "" {
		return
	}

	resp, err := s.dockerManager.ServiceTaskList(ctx, data.Service.ID,
		[]swarm.TaskState{swarm.TaskStateRunning})
	if err != nil {
		return
	}
	nodes := driftingNodes(resp.Items, pinnedNodeID)
	if len(nodes) == 0 {
		return
	}
	s.logger.Warnf(ctx,
		"app %s runs on %v but mounts a volume pinned to %s; applying the pin moves it, "+
			"and the data it has been using stays where it is",
		data.Service.Spec.Name, nodes, pinnedNodeID)
}
```

Add imports `context`, `swarm`, and `placementservice` to `pin_drift.go`. Check `s.logger.Warnf`'s exact signature against another caller (`grep -rn "logger.Warnf" hivepaas_app/service | head -3`) and match it.

- [ ] **Step 6: Verify and commit**

Run: `go build ./... && go test ./hivepaas_app/service/placementservice/... -v && make lint-local`

```bash
git add hivepaas_app/service/placementservice/
git commit -m "Warn when a volume pin would move a running task"
```

---

### Task 10: Confirm deletion needs no new mechanism

The spec assumed deleting a volume would need new per-node cleanup. It does not: `clustercleanupservice.cleanupVolumes` already runs `VolumePrune` on **every** node through the agent (`usecaseagent/nodecleanupagentuc`), and a volume whose setting is gone is mounted by no service, so it is exactly what prune reclaims.

This task verifies that rather than building a second mechanism, and records what was found.

**Files:**
- Modify: `docs/superpowers/specs/2026-09-11-volume-lazy-materialization-design.md` (the "Deletion" section)

**Interfaces:**
- Consumes: nothing.
- Produces: nothing.

- [ ] **Step 1: Confirm the prune reaches every node**

Run:

```bash
grep -rn "VolumePrune" hivepaas_app --include='*.go' | grep -v _test
grep -rn "NodeCleanup" hivepaas_app/usecaseagent/nodecleanupagentuc/*.go
```

Expected: `cleanupVolumes` calls `VolumePrune`, and the agent usecase is what runs a cleanup on each node. If either is not true, stop and report — the spec's deletion section then needs revisiting rather than deleting.

- [ ] **Step 2: Confirm deletion tolerates a volume that was never materialized**

Read `hivepaas_app/usecase/cluster/volumeuc/delete.go`. `VolumeRemove` is already called with the not-found case ignored:

```go
_, err := uc.dockerManager.VolumeRemove(ctx, data.Setting.RefID, true)
if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
	return hperrors.Wrap(err)
}
```

With lazy materialization the manager node frequently has no copy, which this already handles. No change.

- [ ] **Step 3: Replace the spec's Deletion section**

Replace the `### 8. Deletion` section of the design doc with:

```markdown
### 8. Deletion

A volume can now exist on several nodes, but no new cleanup is needed for that.
`clustercleanupservice` already prunes unused volumes on every node through the
agent, and a volume whose setting is gone is mounted by no service - which is
precisely what prune reclaims. `DeleteVolume` keeps removing the manager node's
copy and already ignores a not-found, which is the common case once volumes are
materialized lazily.

One pre-existing hazard is worth recording, because it is easy to mistake for
something this change introduced: the prune is unconditional, so a volume
belonging to a service scaled to zero is unused and can be reclaimed. For bind
volumes that costs nothing - the data is in the bind target, not in the volume -
but a plain local volume with no bind, nfs or tmpfs options keeps its data in
`/var/lib/docker/volumes` and would lose it. Lazy materialization makes this
strictly less likely rather than more, since a volume nothing has mounted does
not exist to be pruned.
```

- [ ] **Step 4: Commit**

```bash
git add docs/superpowers/specs/2026-09-11-volume-lazy-materialization-design.md
git commit -m "Record that volume deletion needs no new cleanup mechanism"
```

---

### Task 11: Drop pinning from the dashboard's update form

**Files** (in `/Users/tnt/go/src/github.com/hivepaas/hivepaas-dashboard`):
- Modify: `src/application/modules/cluster/module-shared/components/volume-form-routes/volume-form-route.helpers.ts`
- Modify: `src/application/modules/cluster/module-shared/components/volume-form/create-or-edit-volume.form.com.tsx`
- Modify: `src/application/modules/cluster/module-shared/components/volume-form-routes/edit-volume-form-route.com.tsx`
- Modify: `src/application/modules/cluster/domain/volumes/volume.entity.ts`

**Interfaces:**
- Consumes: Task 2's update request shape.
- Produces: `toVolumeUpdatePayload(values, updateVer)` returning `{ updateVer, inheritable, default }` only.

- [ ] **Step 1: Remove the pair from the update payload**

In `volume-form-route.helpers.ts`, reduce `toVolumeUpdatePayload` to:

```ts
export function toVolumeUpdatePayload(
    values: CreateOrEditVolumeFormOutput,
    updateVer: number,
): ClusterVolumeUpdatePayload {
    return {
        updateVer,
        inheritable: values.inheritable,
        default: values.default,
    };
}
```

In `volume.entity.ts`, remove `nodeId` and `nodeLabel` from `ClusterVolumeUpdatePayload`, leaving them on `ClusterVolume` and `ClusterVolumeBasePayload`.

- [ ] **Step 2: Make the pinning fieldset read-only on update**

In `create-or-edit-volume.form.com.tsx`, delete `PINNING_CHANGE_WARNING`, the `warnOnPinningChange` prop, the `pinningMoved` computation and the warning block it rendered. Change the pinning fieldset's disabled expression to include the core flag, so pinning is fixed once the volume exists:

```tsx
const pinningDisabled = readOnlyCore || readOnlyInherited || readOnlyPermission || isPending;
```

- [ ] **Step 3: Stop passing the removed prop**

In `edit-volume-form-route.com.tsx`, delete the `warnOnPinningChange` line.

- [ ] **Step 4: Verify**

```bash
cd /Users/tnt/go/src/github.com/hivepaas/hivepaas-dashboard
npx tsc --noEmit 2>&1 | tail -5
npx eslint src/application/modules/cluster/module-shared/components/volume-form src/application/modules/cluster/module-shared/components/volume-form-routes src/application/modules/cluster/domain/volumes
```

Expected: the same 37 pre-existing `tsc` errors as before this work, none in volume files; eslint clean.

- [ ] **Step 5: Commit**

```bash
git add src/application/modules/cluster/
git commit -m "Fix volume node pinning at creation in the dashboard"
```

---

## Final verification

- [ ] **Full suite**

```bash
cd /Users/tnt/go/src/github.com/hivepaas/hivepaas
make test && make lint-local
```

Expected: all packages pass, no lint findings.

- [ ] **Manual check on a live swarm**

Create a bind volume pinned to a node through the API, mount it on an app, deploy, then confirm on that node:

```bash
docker volume inspect <setting-ulid> --format '{{.Options}}'
docker service inspect <service> --format '{{json .Spec.TaskTemplate.Placement.Constraints}}'
```

Expected: the volume carries the recorded options, and the constraint names the pinned node. Confirm too that the volume does **not** exist anywhere before the first task mounts it.

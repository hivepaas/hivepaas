package volumeserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/services/docker"
)

const testCurrentNodeID = "node-manager"

// fakeDockerManager embeds the interface so only the two methods this path
// reaches need a body; anything else would panic loudly rather than pass.
type fakeDockerManager struct {
	docker.Manager
	volumes []volume.Volume
}

func (f *fakeDockerManager) VolumeList(
	_ context.Context, _ ...docker.VolumeListOption,
) (*client.VolumeListResult, error) {
	return &client.VolumeListResult{Items: f.volumes}, nil
}

func (f *fakeDockerManager) NodeCurrentID(_ context.Context) (string, error) {
	return testCurrentNodeID, nil
}

// fakeSettingRepo hands back the stored settings and keeps whatever was written.
type fakeSettingRepo struct {
	repository.SettingRepo
	stored   []*entity.Setting
	upserted []*entity.Setting
}

func (f *fakeSettingRepo) List(
	_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ *basedto.Paging,
	_ ...bunex.SelectQueryOption,
) ([]*entity.Setting, *basedto.PagingMeta, error) {
	return f.stored, &basedto.PagingMeta{}, nil
}

func (f *fakeSettingRepo) UpsertMulti(
	_ context.Context, _ database.IDB, settings []*entity.Setting,
	_, _ []string, _ ...bunex.InsertQueryOption,
) error {
	f.upserted = settings
	return nil
}

func newSyncTest(volumes []volume.Volume, stored []*entity.Setting) (*service, *fakeSettingRepo) {
	repo := &fakeSettingRepo{stored: stored}
	return &service{
		dockerManager: &fakeDockerManager{volumes: volumes},
		settingRepo:   repo,
	}, repo
}

func localVolume(name string) volume.Volume {
	return volume.Volume{Name: name, Driver: "local"}
}

func clusterVolume(name string) volume.Volume {
	return volume.Volume{
		Name:          name,
		Driver:        "some-csi-driver",
		ClusterVolume: &volume.ClusterVolume{ID: "csi-" + name},
	}
}

// storedVolume is a volume HivePaaS already knows about, pinned however the
// operator left it. It is built the way a row read back from the database
// looks - through SetData, then handed back with no parsed cache - because
// MustSetData on the returned setting would leave sync reading the in-memory
// struct instead of something shaped like Data actually is.
func storedVolume(name string, pinning *entity.ClusterVolume) *entity.Setting {
	setting := &entity.Setting{
		ID:    "setting-" + name,
		Type:  base.SettingTypeClusterVolume,
		Kind:  "local",
		Name:  name,
		RefID: name,
	}
	setting.MustSetData(pinning)
	return &entity.Setting{
		ID:    setting.ID,
		Type:  setting.Type,
		Kind:  setting.Kind,
		Name:  setting.Name,
		RefID: setting.RefID,
		Data:  setting.Data,
	}
}

func upsertedByName(repo *fakeSettingRepo, name string) *entity.Setting {
	for _, setting := range repo.upserted {
		if setting.Name == name {
			return setting
		}
	}
	return nil
}

// The guess for a volume nobody has recorded yet: the daemon that listed it is
// the node it is on.
func TestSyncVolumesPinsANewlyDiscoveredVolume(t *testing.T) {
	uc, repo := newSyncTest([]volume.Volume{localVolume("data")}, nil)

	_, err := uc.SyncVolumes(context.Background(), nil)
	assert.NoError(t, err)

	setting := upsertedByName(repo, "data")
	assert.NotNil(t, setting)
	// Read back out of the stored JSON, not the parsed struct: Data is what the
	// repository writes, and a pinning that only exists in memory is no pinning.
	assert.Contains(t, setting.Data, `"nodeId":"`+testCurrentNodeID+`"`)
	assert.Positive(t, setting.Size, "Size has to follow the payload")

	parsed, err := setting.AsClusterVolume()
	assert.NoError(t, err)
	assert.Equal(t, testCurrentNodeID, parsed.NodeID)
}

// A swarm cluster volume is on no node in particular, and says so.
func TestSyncVolumesLeavesAClusterVolumeUnpinned(t *testing.T) {
	uc, repo := newSyncTest([]volume.Volume{clusterVolume("shared")}, nil)

	_, err := uc.SyncVolumes(context.Background(), nil)
	assert.NoError(t, err)

	setting := upsertedByName(repo, "shared")
	assert.NotNil(t, setting)
	assert.NotContains(t, setting.Data, "nodeId")

	parsed, err := setting.AsClusterVolume()
	assert.NoError(t, err)
	assert.Empty(t, parsed.NodeID)
	assert.Empty(t, parsed.NodeLabel)
}

// The point of the rule: an empty pinning on a volume already recorded is the
// operator saying the path is the same on every node - a bind onto shared
// storage looks exactly like a local one, so a sync that filled it in would
// overwrite that answer every pass and the operator could never make it stick.
func TestSyncVolumesNeverOverwritesRecordedPinning(t *testing.T) {
	// Each fixture carries a recorded specification so the new backfill has
	// nothing to add - otherwise this test would fail for the wrong reason,
	// with the write coming from the specification rather than the pinning.
	recordedSpec := func(pin entity.ClusterVolume) *entity.ClusterVolume {
		pin.Managed = true
		pin.Driver = "local"
		pin.DriverOpts = map[string]string{"device": "/srv/data"}
		return &pin
	}

	tests := []struct {
		name    string
		pinning *entity.ClusterVolume
	}{
		{"left empty on purpose - shared storage", recordedSpec(entity.ClusterVolume{})},
		{"pinned by label", recordedSpec(entity.ClusterVolume{NodeLabel: "storage=fast"})},
		{"pinned to another node", recordedSpec(entity.ClusterVolume{NodeID: "node-other"})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stored := storedVolume("data", tt.pinning)
			uc, repo := newSyncTest([]volume.Volume{localVolume("data")}, []*entity.Setting{stored})

			_, err := uc.SyncVolumes(context.Background(), nil)
			assert.NoError(t, err)

			assert.Empty(t, repo.upserted, "nothing docker is authoritative about moved")

			parsed, err := stored.AsClusterVolume()
			assert.NoError(t, err)
			assert.Equal(t, tt.pinning.NodeID, parsed.NodeID)
			assert.Equal(t, tt.pinning.NodeLabel, parsed.NodeLabel)
		})
	}
}

// What docker is authoritative about is still carried over, and that is what
// decides whether the row is written at all.
func TestSyncVolumesCarriesOverWhatDockerOwns(t *testing.T) {
	stored := storedVolume("data", &entity.ClusterVolume{NodeID: "node-other"})
	stored.Kind = "stale-driver"

	uc, repo := newSyncTest([]volume.Volume{localVolume("data")}, []*entity.Setting{stored})

	_, err := uc.SyncVolumes(context.Background(), nil)
	assert.NoError(t, err)

	setting := upsertedByName(repo, "data")
	assert.NotNil(t, setting, "the driver changed, so the row is written")
	assert.Equal(t, "local", setting.Kind)

	parsed, err := setting.AsClusterVolume()
	assert.NoError(t, err)
	assert.Equal(t, "node-other", parsed.NodeID, "still the operator's answer")
}

// A managed setting whose volume has not materialized anywhere is left alone:
// that is the normal state of a volume before any task has mounted it, not
// evidence it was removed. Only HivePaaS's own delete path marks a volume
// deleted.
func TestSyncVolumesKeepsUnmaterializedSettings(t *testing.T) {
	stored := storedVolume("data", &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"device": "/srv/pgdata"},
	})

	uc, repo := newSyncTest(nil, []*entity.Setting{stored})

	_, err := uc.SyncVolumes(context.Background(), nil)
	assert.NoError(t, err)

	assert.Empty(t, repo.upserted, "nothing docker said, so nothing to write")
	assert.True(t, stored.DeletedAt.IsZero(), "absence is not deletion")
}

// A setting recorded before Driver/DriverOpts/Labels existed gets them filled
// in from what docker already knows about the volume, without disturbing the
// pin the operator already gave it.
func TestSyncVolumesBackfillsTheSpecification(t *testing.T) {
	stored := storedVolume("data", &entity.ClusterVolume{NodeID: "node-1"})

	vol := localVolume("data")
	vol.Options = map[string]string{"type": "none", "device": "/srv/pgdata", "o": "bind,rw"}

	uc, repo := newSyncTest([]volume.Volume{vol}, []*entity.Setting{stored})

	_, err := uc.SyncVolumes(context.Background(), nil)
	assert.NoError(t, err)

	setting := upsertedByName(repo, "data")
	assert.NotNil(t, setting)
	assert.Contains(t, setting.Data, `"managed":true`)
	assert.Contains(t, setting.Data, `"driver":"local"`)
	assert.Contains(t, setting.Data, `"device":"/srv/pgdata"`)
	assert.Contains(t, setting.Data, `"nodeId":"node-1"`, "the pin survives the backfill")
}

// Once a specification is recorded it is as much the operator's answer as the
// pin: a device path that moved in docker does not get to overwrite the one
// already on file.
func TestSyncVolumesNeverOverwritesARecordedSpecification(t *testing.T) {
	stored := storedVolume("data", &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"device": "/srv/original"},
	})

	vol := localVolume("data")
	vol.Options = map[string]string{"device": "/srv/moved"}

	uc, repo := newSyncTest([]volume.Volume{vol}, []*entity.Setting{stored})

	_, err := uc.SyncVolumes(context.Background(), nil)
	assert.NoError(t, err)

	assert.Empty(t, repo.upserted, "the recorded specification is not touched")

	parsed, err := stored.AsClusterVolume()
	assert.NoError(t, err)
	assert.Equal(t, "/srv/original", parsed.DriverOpts["device"])
}

// A stored setting for a swarm cluster volume is never claimed as managed -
// swarm owns cluster volumes, and backfillVolumeSpec leaves them alone
// entirely.
func TestSyncVolumesDoesNotClaimClusterVolumes(t *testing.T) {
	stored := storedVolume("shared", &entity.ClusterVolume{})
	stored.RefID = "csi-shared"
	stored.Kind = "some-csi-driver"

	uc, repo := newSyncTest([]volume.Volume{clusterVolume("shared")}, []*entity.Setting{stored})

	_, err := uc.SyncVolumes(context.Background(), nil)
	assert.NoError(t, err)

	assert.Empty(t, repo.upserted, "nothing changed, so nothing to write")

	parsed, err := stored.AsClusterVolume()
	assert.NoError(t, err)
	assert.False(t, parsed.Managed)
	assert.Empty(t, parsed.Driver)
}

// The regression the name carry-over used to cause: docker's name for a
// lazily-created volume is a ULID, recorded in RefID, and it must never
// overwrite the name an operator chose.
func TestSyncVolumesKeepsTheChosenName(t *testing.T) {
	stored := storedVolume("pgdata", &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"device": "/srv/pgdata"},
	})
	stored.RefID = "01HXAMPLEULIDFORTHISVOLUME0"

	uc, repo := newSyncTest([]volume.Volume{localVolume(stored.RefID)}, []*entity.Setting{stored})

	_, err := uc.SyncVolumes(context.Background(), nil)
	assert.NoError(t, err)

	assert.Empty(t, repo.upserted, "nothing docker owns changed")
	assert.Equal(t, "pgdata", stored.Name)
}

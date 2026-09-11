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
// operator left it.
func storedVolume(name string, pinning *entity.ClusterVolume) *entity.Setting {
	setting := &entity.Setting{
		ID:    "setting-" + name,
		Type:  base.SettingTypeClusterVolume,
		Kind:  "local",
		Name:  name,
		RefID: name,
	}
	setting.MustSetData(pinning)
	return setting
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
	tests := []struct {
		name    string
		pinning *entity.ClusterVolume
	}{
		{"left empty on purpose - shared storage", &entity.ClusterVolume{}},
		{"pinned by label", &entity.ClusterVolume{NodeLabel: "storage=fast"}},
		{"pinned to another node", &entity.ClusterVolume{NodeID: "node-other"}},
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

// A volume that disappeared from docker is marked deleted, unchanged by the rest.
func TestSyncVolumesMarksVanishedVolumesDeleted(t *testing.T) {
	stored := storedVolume("gone", &entity.ClusterVolume{NodeID: "node-other"})

	uc, repo := newSyncTest(nil, []*entity.Setting{stored})

	_, err := uc.SyncVolumes(context.Background(), nil)
	assert.NoError(t, err)

	setting := upsertedByName(repo, "gone")
	assert.NotNil(t, setting)
	assert.False(t, setting.DeletedAt.IsZero())
}

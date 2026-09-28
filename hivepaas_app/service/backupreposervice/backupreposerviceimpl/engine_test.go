package backupreposerviceimpl

import (
	"context"
	"errors"
	"testing"

	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
)

// newVolumeSetting builds the volume setting the way a row read back from the
// database looks: SetData caches the parsed struct on the receiver, so a setting
// that still carries it would prove nothing about what AsClusterVolume reads.
// Only the serialized Data travels from storage, so only that is handed on.
func newVolumeSetting(t *testing.T, id, refID, name string, vol *entity.ClusterVolume) *entity.Setting {
	t.Helper()

	setting := &entity.Setting{ID: id, RefID: refID, Type: base.SettingTypeClusterVolume, Name: name}
	assert.NoError(t, setting.SetData(vol))
	return &entity.Setting{
		ID: setting.ID, RefID: setting.RefID, Name: setting.Name, Type: setting.Type, Data: setting.Data,
	}
}

// recordingDockerManager answers one VolumeInspect and remembers what it was
// asked for. The embedded nil interface makes every other method a panic, which
// is the point: a test that reaches docker for anything else should fail loudly.
type recordingDockerManager struct {
	docker.Manager

	inspected  []string
	inspectRes *client.VolumeInspectResult
	inspectErr error
	// currentNode is the node this process's daemon is on.
	currentNode string
}

func (m *recordingDockerManager) NodeCurrentID(context.Context) (string, error) {
	if m.currentNode == "" {
		return "", errors.New("docker daemon is not part of a swarm")
	}
	return m.currentNode, nil
}

func (m *recordingDockerManager) VolumeInspect(
	_ context.Context,
	volumeID string,
	_ ...docker.VolumeInspectOption,
) (*client.VolumeInspectResult, error) {
	m.inspected = append(m.inspected, volumeID)
	if m.inspectErr != nil {
		return nil, m.inspectErr
	}
	if m.inspectRes == nil {
		// What the daemon really answers for a volume that was never materialized
		// on it - the state every lazily created volume is in.
		return nil, errors.New("no such volume: " + volumeID)
	}
	return m.inspectRes, nil
}

// The case this regressed on: a bind volume pinned to another node. Its docker
// volume was never created anywhere, least of all on the manager this process
// talks to, so any inspect would fail and take every backup and restore on the
// repository with it. The setting already knows where the data is.
func TestBuildLocalStorageUsesTheRecordedDeviceWithoutAskingDocker(t *testing.T) {
	dockerManager := &recordingDockerManager{}
	s := &service{dockerManager: dockerManager}

	setting := newVolumeSetting(t, "vol-backups", "01JVOLULID", "backups", &entity.ClusterVolume{
		NodeID:     "node-2",
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "none", "device": "/srv/backups", "o": "bind"},
	})
	repo := &entity.BackupRepo{
		Volume:        entity.ObjectID{ID: "vol-backups"},
		StoragePrefix: "/team-a/",
	}
	refObjects := &entity.RefObjects{RefSettings: map[string]*entity.Setting{"vol-backups": setting}}

	storage, err := s.buildLocalStorage(context.Background(), repo, refObjects)
	if err != nil {
		t.Fatalf("buildLocalStorage failed: %v", err)
	}
	assert.Equal(t, "/host/srv/backups/team-a", storage.Path)
	assert.Equal(t, "node-2", storage.NodeID)
	assert.Empty(t, dockerManager.inspected, "the setting answers this; docker must not be consulted")
}

// A volume with no device recorded - discovered before backfill ran, or a plain
// local volume whose data is under the daemon's own volume root - still has to
// be inspected, and by the docker-side identity rather than the human name.
func TestResolveVolumeHostPathInspectsByRefIDWhenNoDeviceIsRecorded(t *testing.T) {
	dockerManager := &recordingDockerManager{
		inspectRes: &client.VolumeInspectResult{
			Volume: volume.Volume{Mountpoint: "/var/lib/docker/volumes/01JVOLULID/_data"},
		},
	}
	s := &service{dockerManager: dockerManager}

	setting := newVolumeSetting(t, "vol-backups", "01JVOLULID", "backups", &entity.ClusterVolume{
		NodeID: "node-1",
	})
	clusterVolume, err := setting.AsClusterVolume()
	assert.NoError(t, err)

	hostPath, err := s.resolveVolumeHostPath(context.Background(), setting, clusterVolume)

	assert.NoError(t, err)
	assert.Equal(t, "/var/lib/docker/volumes/01JVOLULID/_data", hostPath)
	assert.Equal(t, []string{"01JVOLULID"}, dockerManager.inspected)
}

// A volume on all nodes is a directory every node reaches at the same path -
// shared storage mounted alike everywhere. A repository on it is reached from any
// node; its commands go to the node HivePaaS runs on.
func TestBuildLocalStorageTakesASharedBindVolumeOnTheCurrentNode(t *testing.T) {
	dockerManager := &recordingDockerManager{currentNode: "node-mgr"}
	s := &service{dockerManager: dockerManager}

	setting := newVolumeSetting(t, "vol-shared", "01JVOLSHARED", "shared-vol", &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "none", "device": "/mnt/nfs/backups", "o": "bind,rw"},
	})
	repo := &entity.BackupRepo{Volume: entity.ObjectID{ID: "vol-shared"}, StoragePrefix: "kopia"}
	refObjects := &entity.RefObjects{RefSettings: map[string]*entity.Setting{"vol-shared": setting}}

	storage, err := s.buildLocalStorage(context.Background(), repo, refObjects)

	assert.NoError(t, err)
	assert.Equal(t, "/host/mnt/nfs/backups/kopia", storage.Path)
	assert.Equal(t, "node-mgr", storage.NodeID)
	assert.Empty(t, storage.NodeLabel)
	assert.True(t, storage.Shared)
	assert.Empty(t, dockerManager.inspected, "the setting answers this; docker must not be consulted")
}

// A volume on all nodes that is not a bind directory is somewhere else on each
// node, or not on the host at all: a docker-managed local volume, an NFS volume
// the driver mounts itself. No path works everywhere.
func TestBuildLocalStorageRefusesASharedVolumeThatIsNotABindDirectory(t *testing.T) {
	cases := map[string]map[string]string{
		"a docker-managed local volume":  nil,
		"a volume the NFS driver mounts": {"type": "nfs", "o": "addr=10.0.0.5,rw", "device": ":/exports/backups"},
	}
	for name, opts := range cases {
		s := &service{dockerManager: &recordingDockerManager{currentNode: "node-mgr"}}
		setting := newVolumeSetting(t, "vol-shared", "01JVOLSHARED", "shared-vol", &entity.ClusterVolume{
			Managed: true, Driver: "local", DriverOpts: opts,
		})
		repo := &entity.BackupRepo{Volume: entity.ObjectID{ID: "vol-shared"}}
		refObjects := &entity.RefObjects{RefSettings: map[string]*entity.Setting{"vol-shared": setting}}

		_, err := s.buildLocalStorage(context.Background(), repo, refObjects)

		assert.ErrorIs(t, err, hperrors.ErrBackupVolumeSharedNotBind, name)
	}
}

// A volume a data backup reads, on all nodes: read on the node HivePaaS runs on.
func TestVolumeHostDirOfASharedVolumeIsOnTheCurrentNode(t *testing.T) {
	s := &service{dockerManager: &recordingDockerManager{currentNode: "node-mgr"}}
	volume := newVolumeSetting(t, "vol-shared", "01JVOLSHARED", "shared-vol", &entity.ClusterVolume{
		Managed:    true,
		Driver:     "local",
		DriverOpts: map[string]string{"type": "none", "device": "/mnt/nfs/data", "o": "bind"},
	})

	dir, err := s.VolumeHostDir(context.Background(), volume)

	assert.NoError(t, err)
	assert.Equal(t, "/mnt/nfs/data", dir.Dir)
	assert.Equal(t, "node-mgr", dir.NodeID)
	assert.Empty(t, dir.NodeLabel)
}

func TestVolumeHostDirRefusesASharedVolumeThatIsNotABindDirectory(t *testing.T) {
	s := &service{dockerManager: &recordingDockerManager{currentNode: "node-mgr"}}
	volume := newVolumeSetting(t, "vol-shared", "01JVOLSHARED", "shared-vol", &entity.ClusterVolume{
		Managed: true, Driver: "local",
	})

	_, err := s.VolumeHostDir(context.Background(), volume)

	assert.ErrorIs(t, err, hperrors.ErrBackupVolumeSharedNotBind)
}

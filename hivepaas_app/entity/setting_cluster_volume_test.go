package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// newStoredClusterVolumeSetting builds a setting and returns it the way a row
// read back from the database looks, with no parsed cache. This ensures the
// specification survives the trip through Data (JSON serialization).
func newStoredClusterVolumeSetting(t *testing.T, vol *ClusterVolume) *Setting {
	t.Helper()

	setting := &Setting{Type: base.SettingTypeClusterVolume}
	assert.NoError(t, setting.SetData(vol))
	// Return a fresh Setting with only Type and Data, simulating a database read.
	return &Setting{Type: setting.Type, Data: setting.Data}
}

// The specification has to survive the trip through Data, because Data is what
// the repository writes and a field that only exists in memory is not recorded.
func TestClusterVolumeSpecRoundTrips(t *testing.T) {
	setting := newStoredClusterVolumeSetting(t, &ClusterVolume{
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
	assert.Equal(t, map[string]string{"type": "none", "device": "/data/pg", "o": "bind,rw"}, parsed.DriverOpts)
	assert.Equal(t, map[string]string{"hivepaas.volume.name": "pgdata"}, parsed.Labels)
}

// A volume recorded before this field existed reads back as unmanaged, which is
// what keeps it mounted by name until a sync backfills its specification.
func TestClusterVolumeDefaultsToUnmanaged(t *testing.T) {
	setting := newStoredClusterVolumeSetting(t, &ClusterVolume{NodeID: "node-1"})

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

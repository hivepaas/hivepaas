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

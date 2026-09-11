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
		TmpfsOptions: &VolumeTmpfsOptionsReq{Size: 100 * unit.MB, UID: 1000},
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

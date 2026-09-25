package volumedto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/docker"
)

func localReq() *VolumeBaseReq {
	return &VolumeBaseReq{Name: "data", Driver: docker.VolumeDriverLocal}
}

// A volume of docker's own storage reaches nothing of the node, and a bind
// whose directory HivePaaS chooses reaches only the storage HivePaaS hands out.
func TestHostAccessOfVolumesThatReachNoPathOfTheirOwn(t *testing.T) {
	plain := localReq().HostAccess()
	assert.False(t, plain.NamesHostPath)
	assert.False(t, plain.ReachesDockerSocket)

	managed := localReq()
	managed.BindOptions = &VolumeBindOptionsReq{}
	assert.False(t, managed.HostAccess().NamesHostPath, "HivePaaS chooses the directory")

	nfs := localReq()
	nfs.NfsOptions = &VolumeNfsOptionsReq{Addr: "10.0.0.1", Device: ":/exports/data"}
	assert.False(t, nfs.HostAccess().NamesHostPath, "a remote server is not this node")

	tmpfs := localReq()
	tmpfs.TmpfsOptions = &VolumeTmpfsOptionsReq{}
	assert.False(t, tmpfs.HostAccess().NamesHostPath)
}

// A directory of the caller's own, however it is described, is a mount of the
// node.
func TestHostAccessOfVolumesThatNameAPathOfTheNode(t *testing.T) {
	bind := localReq()
	bind.BindOptions = &VolumeBindOptionsReq{Directory: "/srv/backups"}
	access := bind.HostAccess()
	assert.True(t, access.NamesHostPath)
	assert.Equal(t, "/srv/backups", access.Directory)
	assert.False(t, access.ReachesDockerSocket)

	raw := localReq()
	raw.Options = map[string]string{"type": "none", "o": "bind,rw", "device": "/srv/backups"}
	access = raw.HostAccess()
	assert.True(t, access.NamesHostPath, "raw driver options describe the same mount")
	assert.Equal(t, "/srv/backups", access.Directory)

	// Half a description still asks for one.
	partial := localReq()
	partial.Options = map[string]string{"type": "none"}
	assert.True(t, partial.HostAccess().NamesHostPath)
}

// The socket, and any directory holding it, is the Docker API of the node.
func TestHostAccessSeesTheDockerSocket(t *testing.T) {
	for _, directory := range []string{"/var/run/docker.sock", "/var/run", "/"} {
		bind := localReq()
		bind.BindOptions = &VolumeBindOptionsReq{Directory: directory}
		assert.True(t, bind.HostAccess().ReachesDockerSocket, directory)

		raw := localReq()
		raw.Options = map[string]string{"device": directory}
		assert.True(t, raw.HostAccess().ReachesDockerSocket, directory)
	}
}

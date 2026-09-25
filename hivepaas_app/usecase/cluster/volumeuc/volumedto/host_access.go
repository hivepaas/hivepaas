package volumedto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// HostAccess is what a volume request asks of the nodes it will be mounted on.
//
// A volume is ordinarily docker's own storage, which reaches nothing of the
// node. Two things change that: a bind directory the caller named, and driver
// options that describe a mount of the node's filesystem. Both are the same
// decision as mounting a path of the host into an app, so both take the same
// permission.
type HostAccess struct {
	// Directory is the path of the node the caller named, empty when it named
	// none. A bind volume whose directory HivePaaS chooses is not this: the
	// directory is then HivePaaS's own storage, given out per scope.
	Directory string
	// NamesHostPath is true when the request describes a mount of the node's
	// filesystem, whether through the directory or through raw driver options.
	NamesHostPath bool
	// ReachesDockerSocket is true when the path would hold the docker socket.
	ReachesDockerSocket bool
}

// HostAccess reads what the request asks of the node. It is answered from the
// request alone, before anything touches a host, so that a refusal creates
// nothing.
func (req *VolumeBaseReq) HostAccess() HostAccess {
	out := HostAccess{}
	if req == nil {
		return out
	}
	if req.Driver == docker.VolumeDriverLocal && req.BindOptions != nil {
		out.Directory = req.BindOptions.Directory
	}
	// Raw options go to the driver as they are, so they describe the mount as
	// fully as the typed fields do - and name their own device.
	if device := req.Options[volumeservice.DriverOptDevice]; device != "" {
		out.Directory = device
	}
	out.NamesHostPath = out.Directory != "" || volumeservice.DriverOptsNameHostPath(req.Options)
	out.ReachesDockerSocket = volumeservice.ReachesDockerSocket(out.Directory)
	return out
}

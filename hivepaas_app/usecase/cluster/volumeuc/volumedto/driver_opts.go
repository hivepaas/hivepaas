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
	// nothing they recognize.
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

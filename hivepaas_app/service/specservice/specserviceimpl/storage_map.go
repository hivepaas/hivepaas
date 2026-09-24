package specserviceimpl

import (
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

// volumeRef is how the bundle names the volume a managed mount reaches: its
// scope path when the export holds the volume, an external reference when it
// does not. Both empty means the volume cannot be named at all.
type volumeRef func(volumeID string) (path string, external *specmodel.ExternalRef)

// mapAppStorage turns an app's mounts into its storage block.
//
// A mount into a volume's directory for an app of this environment is written
// the way a template and the storage screen write it: the volume, the path below
// that app's directory, and the app when it is not this one. That is the form
// import can build again anywhere - Docker's own form carries a host path or a
// volume name that means nothing on another installation. Every other mount is
// kept as Docker holds it. The shared-memory mount is neither:
// resources.memory.shmSize carries it.
//
// descs is volumeservice.DescribeAppMounts' answer for the same mounts, in the
// same order.
func mapAppStorage(
	task *swarm.TaskSpec,
	descs []*volumeservice.AppMountDesc,
	ref volumeRef,
) (*specmodel.Storage, error) {
	if task.ContainerSpec == nil || len(task.ContainerSpec.Mounts) == 0 {
		return nil, nil
	}
	mounts := task.ContainerSpec.Mounts
	shm := dockerhelper.GetShmMount(task)

	out := &specmodel.Storage{}
	seen := make(map[string]bool, len(mounts))
	for i := range mounts {
		m := &mounts[i]
		if shm != nil && m.Type == shm.Type && m.Target == shm.Target {
			continue
		}
		// The socket is this installation's, named after this app's id; import
		// gives an app with the setting its own.
		if dockerapiservice.IsSocketMount(m) {
			continue
		}
		if seen[m.Target] {
			return nil, hperrors.Wrap(hperrors.ErrSpecMountTargetDuplicated).WithParam("Target", m.Target)
		}
		seen[m.Target] = true

		var desc *volumeservice.AppMountDesc
		if i < len(descs) {
			desc = descs[i]
		}
		if managed, ok := mapManagedMount(m, desc, ref); ok {
			out.Mounts = withMount(out.Mounts, m.Target, managed)
			continue
		}
		out.DockerMounts = withMount(out.DockerMounts, m.Target, mapMount(m))
	}
	if len(out.Mounts) == 0 && len(out.DockerMounts) == 0 {
		return nil, nil
	}
	return out, nil
}

// mapManagedMount writes a mount the way the storage screen writes it, when it
// is one: a directory an app of this environment has in a volume the bundle can
// name.
func mapManagedMount(
	m *mount.Mount,
	desc *volumeservice.AppMountDesc,
	ref volumeRef,
) (specmodel.Mount, bool) {
	if desc == nil || desc.AppKey == "" || desc.VolumeID == "" {
		return specmodel.Mount{}, false
	}
	path, external := ref(desc.VolumeID)
	if path == "" && external == nil {
		return specmodel.Mount{}, false
	}

	// A managed local volume reaches Docker as a bind, but what was asked for is
	// the volume, and building that again makes the same bind.
	out := specmodel.Mount{
		Type:        mount.TypeVolume,
		Source:      path,
		External:    external,
		Consistency: m.Consistency,
	}
	noCopy := m.VolumeOptions != nil && m.VolumeOptions.NoCopy
	var opts *specmodel.VolumeOptions
	if desc.Subpath != "" || noCopy {
		opts = &specmodel.VolumeOptions{Subpath: desc.Subpath, NoCopy: noCopy}
	}
	if m.Type == mount.TypeCluster {
		out.Type, out.ClusterOptions = mount.TypeCluster, opts
	} else {
		out.VolumeOptions = opts
	}

	if desc.Own {
		out.ReadOnly = m.ReadOnly
	} else {
		// Whether this app may change the other's files is what Write says; the
		// builder derives the mount's read-only flag from it, so it is not
		// written twice.
		out.SourceApp = &specmodel.MountSourceApp{App: desc.AppKey, Write: !m.ReadOnly}
	}
	return out, true
}

func withMount(mounts map[string]specmodel.Mount, target string, m specmodel.Mount) map[string]specmodel.Mount {
	if mounts == nil {
		mounts = map[string]specmodel.Mount{}
	}
	mounts[target] = m
	return mounts
}

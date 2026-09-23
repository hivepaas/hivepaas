package specmodel

import (
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
)

var serviceModes = []docker.ServiceMode{
	docker.ServiceModeReplicated, docker.ServiceModeReplicatedJob,
	docker.ServiceModeGlobal, docker.ServiceModeGlobalJob,
}

// CheckImportable refuses an exported document that cannot be built as it
// stands. It holds none of a template's caps - the document describes what an
// installation already ran - only what building needs: every volume resolved to
// a setting, every mount at an absolute target of its own, a subpath that stays
// below its directory, a mode swarm knows, ports that are ports.
func CheckImportable(doc *AppDoc) error {
	if doc == nil || doc.Deployment == nil {
		return nil
	}
	d := doc.Deployment
	if err := checkImportedStorage(d.Storage); err != nil {
		return err
	}
	if s := d.Service; s != nil && s.ModeSpec != nil && s.ModeSpec.Mode != "" &&
		!slices.Contains(serviceModes, s.ModeSpec.Mode) {
		return invalid("deployment.service.modeSpec.mode: %q is not a mode swarm knows", s.ModeSpec.Mode)
	}
	if n := d.Networks; n != nil && n.EndpointSpec != nil {
		for i, port := range n.EndpointSpec.Ports {
			if port != nil && (port.Target < 1 || port.Target > maxPortNumber || port.Published > maxPortNumber) {
				return invalid("deployment.networks.endpointSpec.ports[%d]: not a port", i)
			}
		}
	}
	return nil
}

func checkImportedStorage(s *Storage) error {
	if s == nil {
		return nil
	}
	for _, target := range slices.Sorted(maps.Keys(s.Mounts)) {
		m, at := s.Mounts[target], "deployment.storage.mounts."+target
		switch {
		case !path.IsAbs(target):
			return invalid("%s: the target is not an absolute path", at)
		case m.External != nil:
			return invalid("%s: the volume %q has to be resolved before the app is built", at, m.External.Name)
		case m.Type != mount.TypeVolume && m.Type != mount.TypeCluster:
			return invalid("%s: a managed mount is a volume or a cluster volume, not %q", at, m.Type)
		case m.Source == "":
			return invalid("%s: no volume", at)
		}
		for _, opts := range []*VolumeOptions{m.VolumeOptions, m.ClusterOptions} {
			if opts != nil && !staysBelow(opts.Subpath) {
				return invalid("%s: the subpath %q leaves the app's directory", at, opts.Subpath)
			}
		}
		if _, twice := s.DockerMounts[target]; twice {
			return invalid("%s: the target is mounted twice", at)
		}
	}
	for _, target := range slices.Sorted(maps.Keys(s.DockerMounts)) {
		if !path.IsAbs(target) {
			return invalid("deployment.storage.dockerMounts.%s: the target is not an absolute path", target)
		}
	}
	return nil
}

// staysBelow reports whether a subpath names a directory below the one it is
// joined to.
func staysBelow(subpath string) bool {
	if subpath == "" {
		return true
	}
	cleaned := path.Clean(subpath)
	return !path.IsAbs(cleaned) && cleaned != ".." && !strings.HasPrefix(cleaned, "../")
}

func invalid(format string, args ...any) error {
	return hperrors.Wrap(hperrors.ErrSpecBlockInvalid).WithExtraDetail("%s", fmt.Sprintf(format, args...))
}

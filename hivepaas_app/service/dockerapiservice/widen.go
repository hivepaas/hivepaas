package dockerapiservice

import (
	"slices"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// Widens reports whether next lets an app do anything prev did not: access
// where it had none, an image, directory, network or group it did not have, or
// a shared volume it did not have or one standing for another directory, or a
// higher limit. Granting that takes what giving access takes; taking any of it
// away does not. nil is no access.
//
// Host mode covers everything the proxy could allow, so entering it widens and
// leaving it never does. Patterns are compared as written, so one that another
// in prev already covers still counts as added. Only "*", every image, is known
// to cover the rest.
func Widens(prev, next *entity.AppDockerAPISettings) bool {
	switch {
	case next == nil, prev.IsHostMode():
		return false
	case prev == nil, next.IsHostMode():
		return true
	}
	if !slices.Contains(prev.Images, "*") && adds(prev.Images, next.Images) {
		return true
	}
	if adds(prev.SharedDirs, next.SharedDirs) || adds(prev.Networks, next.Networks) || adds(prev.Allow, next.Allow) {
		return true
	}
	for name, dir := range next.SharedVolumes {
		if was, found := prev.SharedVolumes[name]; !found || was != dir {
			return true
		}
	}
	was, is := effectiveLimits(prev.Limits), effectiveLimits(next.Limits)
	return is.Containers > was.Containers || is.Memory > was.Memory || is.CPUs > was.CPUs
}

// EntersHostMode reports whether next gives the node's own socket to an app
// that did not have it: from no access, from host mode turned off, or from the
// proxy. That takes the privileged-apps switch and an administrator.
func EntersHostMode(prev, next *entity.AppDockerAPISettings) bool {
	return next.IsHostMode() && !prev.IsHostMode()
}

// adds reports whether next holds a value prev does not.
func adds(prev, next []string) bool {
	return slices.ContainsFunc(next, func(value string) bool { return !slices.Contains(prev, value) })
}

// effectiveLimits are limits with each zero replaced by the default it stands
// for.
func effectiveLimits(limits entity.AppDockerAPILimits) entity.AppDockerAPILimits {
	return entity.AppDockerAPILimits{
		Containers: gofn.Coalesce(limits.Containers, DefaultContainers),
		Memory:     gofn.Coalesce(limits.Memory, DefaultMemory),
		CPUs:       gofn.Coalesce(limits.CPUs, DefaultCPUs),
	}
}

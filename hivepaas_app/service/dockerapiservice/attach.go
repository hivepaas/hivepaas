package dockerapiservice

import (
	"slices"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
)

// SocketMount is the mount that puts an app's socket where the app looks for it.
func SocketMount(appID string) mount.Mount {
	return mount.Mount{Type: mount.TypeVolume, Source: SocketVolumeName(appID), Target: SocketMountTarget}
}

// IsSocketMount reports a mount of any app's socket volume.
func IsSocketMount(m *mount.Mount) bool {
	return m.Type == mount.TypeVolume && strings.HasPrefix(m.Source, SocketVolumePrefix)
}

// IsAppNetworkName reports the name of any app's own network.
func IsAppNetworkName(name string) bool {
	return strings.HasPrefix(name, NetworkPrefix)
}

func isSocketMount(m mount.Mount) bool {
	return IsSocketMount(&m)
}

// Attach gives a service spec an app's socket and its network, once each. A
// socket of another app - the one a clone was copied from - is replaced.
func Attach(spec *swarm.ServiceSpec, appID, networkID string) {
	container := spec.TaskTemplate.ContainerSpec
	if container == nil {
		container = &swarm.ContainerSpec{}
		spec.TaskTemplate.ContainerSpec = container
	}
	container.Mounts = append(slices.DeleteFunc(container.Mounts, isSocketMount), SocketMount(appID))
	attached := slices.ContainsFunc(spec.TaskTemplate.Networks, func(n swarm.NetworkAttachmentConfig) bool {
		return n.Target == networkID
	})
	if !attached {
		spec.TaskTemplate.Networks = append(spec.TaskTemplate.Networks, swarm.NetworkAttachmentConfig{Target: networkID})
	}
}

// Detach takes every socket mount off a spec, and the attachment to networkID
// when there is one.
func Detach(spec *swarm.ServiceSpec, networkID string) {
	if container := spec.TaskTemplate.ContainerSpec; container != nil {
		container.Mounts = slices.DeleteFunc(container.Mounts, isSocketMount)
	}
	if networkID == "" {
		return
	}
	spec.TaskTemplate.Networks = slices.DeleteFunc(spec.TaskTemplate.Networks,
		func(n swarm.NetworkAttachmentConfig) bool { return n.Target == networkID })
}

// KeepSocketMounts is what a screen that rewrites an app's mounts saves: the
// mounts it built, and the socket the service already has. The storage screen
// does not show the socket, since it is given with the app's access rather than
// chosen there, so what it sends back never has it.
func KeepSocketMounts(final, current []mount.Mount) []mount.Mount {
	kept := slices.DeleteFunc(slices.Clone(final), isSocketMount)
	for _, m := range current {
		if isSocketMount(m) {
			kept = append(kept, m)
		}
	}
	return kept
}

// KeepAppNetwork is the same for the network screen and the app's own network,
// networkID, when the service is attached to it.
func KeepAppNetwork(final, current []swarm.NetworkAttachmentConfig,
	networkID string) []swarm.NetworkAttachmentConfig {
	if networkID == "" {
		return final
	}
	for _, n := range current {
		if n.Target != networkID {
			continue
		}
		kept := slices.DeleteFunc(slices.Clone(final), func(f swarm.NetworkAttachmentConfig) bool {
			return f.Target == networkID
		})
		return append(kept, n)
	}
	return final
}

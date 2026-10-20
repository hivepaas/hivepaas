package appservice

import (
	"slices"

	"github.com/moby/moby/api/types/swarm"
)

// AddNetworkAlias gives spec's service alias on the network it is reached by
// its key on - the project's, which every app is made on - or on its first
// network when none carries the key. It reports whether spec changed.
func AddNetworkAlias(spec *swarm.ServiceSpec, key, alias string) bool {
	networks := spec.TaskTemplate.Networks
	if len(networks) == 0 {
		return false
	}
	at := slices.IndexFunc(networks, func(n swarm.NetworkAttachmentConfig) bool {
		return slices.Contains(n.Aliases, key)
	})
	at = max(at, 0)
	if slices.Contains(networks[at].Aliases, alias) {
		return false
	}
	networks[at].Aliases = append(slices.Clone(networks[at].Aliases), alias)
	return true
}

// RemoveNetworkAlias takes alias off spec's service on every network. It
// reports whether spec changed.
func RemoveNetworkAlias(spec *swarm.ServiceSpec, alias string) bool {
	changed := false
	for i := range spec.TaskTemplate.Networks {
		network := &spec.TaskTemplate.Networks[i]
		if slices.Contains(network.Aliases, alias) {
			network.Aliases = slices.DeleteFunc(slices.Clone(network.Aliases),
				func(one string) bool { return one == alias })
			changed = true
		}
	}
	return changed
}

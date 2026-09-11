package placementservice

import (
	"fmt"
	"strings"
)

// VolumePin is a volume's claim about which node its data is on.
type VolumePin struct {
	VolumeName string
	NodeID     string
	NodeLabel  string
}

func (p VolumePin) isPinned() bool {
	return p.NodeID != "" || p.NodeLabel != ""
}

// constraint is the swarm spelling of this pin, empty when there is none.
func (p VolumePin) constraint() string {
	if p.NodeID != "" {
		return "node.id==" + p.NodeID
	}
	if p.NodeLabel == "" {
		return ""
	}
	key, val, found := strings.Cut(p.NodeLabel, "=")
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	val = strings.TrimSpace(val)
	if !found || val == "" {
		// A bare key means "this label is set", which swarm writes as "true".
		val = "true"
	}
	return fmt.Sprintf("node.labels.%s==%s", key, val)
}

// VolumePinConflict is two volumes that cannot both be reached from one node.
//
//nolint:errname // named for what it holds, not the Xxx0Error convention - callers read it as data
type VolumePinConflict struct {
	First  VolumePin
	Second VolumePin
}

func (c *VolumePinConflict) Error() string {
	return fmt.Sprintf("volume '%s' is pinned to %s and volume '%s' to %s, so no node has both",
		c.First.VolumeName, c.First.constraint(),
		c.Second.VolumeName, c.Second.constraint())
}

// VolumePinConstraint is the single placement constraint the given volumes
// require, or a conflict when they require different ones.
//
// Unlike every other constraint HivePaaS emits this one is required rather than
// an exclusion, so two of them disagreeing is unsatisfiable rather than merely
// narrow - which is why the disagreement is returned instead of applied.
func VolumePinConstraint(pins []VolumePin) (string, *VolumePinConflict) {
	var chosen VolumePin
	for _, pin := range pins {
		if !pin.isPinned() || pin.constraint() == "" {
			continue
		}
		if !chosen.isPinned() {
			chosen = pin
			continue
		}
		if chosen.constraint() != pin.constraint() {
			return "", &VolumePinConflict{First: chosen, Second: pin}
		}
	}
	return chosen.constraint(), nil
}

package placementservice

import (
	"path/filepath"
	"strings"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// VolumePinsForMounts resolves each of a service's mounts back to the volume
// setting it came from and returns that volume's pin.
//
// TypeVolume and TypeCluster mounts name their volume directly - their Source
// is the volume's RefID. A TypeBind mount does not: useBindMountIfAppropriate
// (appsettingsuc/storage_settings_update.go) rewrites a `local` volume whose
// options are `type=none` + `device=<dir>` into a plain host-path bind before
// it ever reaches docker, so the mount that comes back carries a path and
// nothing else. The only way back to the volume is to compare that path
// against the device recorded in the volume's own DriverOpts when it was
// specified.
//
// Every other mount type (tmpfs, npipe, ...) carries no identity a volume
// could be matched against and is skipped, as is a bind whose source matches
// no volume - the user may have added it directly, unrelated to any volume
// HivePaaS knows about.
func VolumePinsForMounts(mounts []mount.Mount, volumes []*entity.Setting) ([]VolumePin, error) {
	clusterVolumes := make([]*entity.ClusterVolume, len(volumes))
	for i, vol := range volumes {
		cv, err := vol.AsClusterVolume()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if cv == nil {
			// A legacy setting with no recorded data at all: still a real volume
			// (matchable by RefID), just one that can never be matched as a bind
			// until something backfills its DriverOpts.
			cv = &entity.ClusterVolume{}
		}
		clusterVolumes[i] = cv
	}

	// matchedAt tracks, in first-seen order, which volumes (by index) at least
	// one mount resolved to - a volume mounted twice must still contribute only
	// one pin.
	var matchedAt []int
	seen := make(map[int]struct{}, len(volumes))
	match := func(i int) {
		if _, ok := seen[i]; ok {
			return
		}
		seen[i] = struct{}{}
		matchedAt = append(matchedAt, i)
	}

	for _, mnt := range mounts {
		switch mnt.Type {
		case mount.TypeVolume, mount.TypeCluster:
			for i, vol := range volumes {
				if vol.RefID == mnt.Source {
					match(i)
				}
			}
		case mount.TypeBind:
			for _, i := range longestDeviceMatches(mnt.Source, clusterVolumes) {
				match(i)
			}
		case mount.TypeTmpfs, mount.TypeNamedPipe, mount.TypeImage:
			// No volume identity to match against.
		}
	}

	pins := make([]VolumePin, 0, len(matchedAt))
	for _, i := range matchedAt {
		pins = append(pins, VolumePin{
			VolumeName: volumes[i].Name,
			NodeID:     clusterVolumes[i].NodeID,
			NodeLabel:  clusterVolumes[i].NodeLabel,
		})
	}
	return pins, nil
}

// longestDeviceMatches finds the cluster volumes whose recorded bind device is
// a path-boundary ancestor of (or equal to) source, and returns only the ones
// with the longest such device - a volume at /srv/data/pg is a more specific,
// and therefore better, answer than one at /srv/data for a source underneath
// both. Both are returned on a tie: an identical device means the two settings
// describe the same directory, and if their pins disagree that is a genuine
// conflict for VolumePinConstraint to report rather than something to guess
// past here.
//
// Only volumes HivePaaS authored are candidates. An unmanaged volume's recorded
// device was copied off somebody else's volume rather than chosen by HivePaaS,
// so it never becomes a bind mount (bindMountTarget refuses it) and no bind
// source can have come from it - matching one would attribute a path, and a
// node pin, to a volume that had nothing to do with it.
func longestDeviceMatches(source string, clusterVolumes []*entity.ClusterVolume) []int {
	source = filepath.Clean(source)

	var best []int
	bestLen := -1
	for i, cv := range clusterVolumes {
		if !cv.Managed || cv.DriverOpts["type"] != "none" {
			continue
		}
		device := cv.DriverOpts["device"]
		if device == "" {
			continue
		}
		device = filepath.Clean(device)
		if source != device && !strings.HasPrefix(source, device+"/") {
			continue
		}

		switch {
		case len(device) > bestLen:
			bestLen = len(device)
			best = []int{i}
		case len(device) == bestLen:
			best = append(best, i)
		}
	}
	return best
}

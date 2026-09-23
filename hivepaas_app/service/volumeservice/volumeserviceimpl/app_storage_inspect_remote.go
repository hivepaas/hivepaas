package volumeserviceimpl

import (
	"bytes"
	"context"
	"path"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// listNonEmptyScript prints back each path it was given that is a directory with
// something in it, and says nothing about the rest.
//
// The paths are positional arguments rather than part of the script, so nothing
// in one is read as shell. They are built from a volume's own device and from
// subpaths safeSubpath has already narrowed to plain names, and this is only a
// read - but a command assembled by string concatenation is the kind of thing
// that stops being only a read when somebody extends it.
const listNonEmptyScript = `for p in "$@"; do
  if [ -d "$p" ] && [ -n "$(ls -A "$p" 2>/dev/null | head -n 1)" ]; then printf '%s\n' "$p"; fi
done`

// inspectVolumeOnItsNode asks the node holding a volume what is in the
// directories, for storage this one cannot open.
//
// The agent sees the whole host filesystem under a fixed prefix, which is how
// the path it is asked about is built. A volume that says nothing about which
// node it is on leaves the states unchecked: there is nobody to ask.
func (s *service) inspectVolumeOnItsNode(
	ctx context.Context,
	setting *entity.Setting,
	states []*volumeservice.AppStorageState,
) {
	vol, err := setting.AsClusterVolume()
	if err != nil || vol == nil {
		return
	}
	if vol.NodeID == "" && vol.NodeLabel == "" {
		return
	}
	device, _, ok := bindMountTarget(vol, "")
	if !ok {
		// A volume with no directory of its own - a driver keeping its data
		// somewhere this cannot name. Nothing to look at by path.
		return
	}

	// Both spellings of the host: the prefix the agent mounts the root at, and
	// the one a runtime that rewrites bind sources serves the same root at. A
	// path is only ever reported back, never acted on, and the answer is matched
	// to the query it was built from.
	candidates := make([]string, 0, len(states)*len(hostPathCandidates))
	owner := make(map[string]*volumeservice.AppStorageState, cap(candidates))
	for _, state := range states {
		for _, candidate := range hostCandidatesFor(path.Join(device, state.Path)) {
			candidates = append(candidates, candidate)
			owner[candidate] = state
		}
	}
	if len(candidates) == 0 {
		return
	}

	stdout := &bytes.Buffer{}
	command := append([]string{"sh", "-c", listNonEmptyScript, "sh"}, candidates...)
	_, err = s.nodeExecService.ExecCommand(ctx, &nodeexecservice.CommandExecReq{
		NodeID:    vol.NodeID,
		NodeLabel: vol.NodeLabel,
		CommandExecOpts: &nodeexecservice.CommandExecOpts{
			Command: command,
			Stdout:  stdout,
		},
	})
	if err != nil {
		// The node could not be asked. Unchecked is what the states already say.
		if s.logger != nil {
			s.logger.Warn("could not read storage on the node holding it",
				"volume", setting.ID, "node", vol.NodeID, "nodeLabel", vol.NodeLabel, "error", err)
		}
		return
	}

	// Every state was asked about, so every one of them has an answer now: the
	// ones named came back with something in them, the rest are empty or absent.
	for _, state := range states {
		state.Checked = true
	}
	for _, line := range strings.Split(stdout.String(), "\n") {
		if state := owner[strings.TrimSpace(line)]; state != nil {
			state.Exists, state.Empty = true, false
		}
	}
}

// hostPathCandidates are the prefixes a host path is reachable at from inside a
// container that mounts the root. The second is what Docker Desktop serves a
// shared directory at, and it is tried only after the plain one.
var hostPathCandidates = []string{"", "/host_mnt"}

func hostCandidatesFor(hostPath string) []string {
	candidates := make([]string, 0, len(hostPathCandidates))
	for _, prefix := range hostPathCandidates {
		candidates = append(candidates, path.Join(volumeservice.HostPathPrefix, prefix, hostPath))
	}
	return candidates
}

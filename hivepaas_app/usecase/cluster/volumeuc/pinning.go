package volumeuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// currentNodeIDMark is what a caller sends instead of a node id when they mean
// "wherever HivePaaS itself is running".
const currentNodeIDMark = "current"

// resolveCurrentNode turns the marker into the id it stands for, and leaves any
// real id alone.
//
// A daemon outside a swarm has no id to give, and NodeCurrentID answers that
// with an error rather than an empty string - which matters here, because empty
// pinning is not "unknown" but the claim that the volume is reachable from every
// node. See docker.NodeCurrentID.
func (uc *UC) resolveCurrentNode(ctx context.Context, nodeID string) (string, error) {
	if nodeID != currentNodeIDMark {
		return nodeID, nil
	}

	resolved, err := uc.dockerManager.NodeCurrentID(ctx)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return resolved, nil
}

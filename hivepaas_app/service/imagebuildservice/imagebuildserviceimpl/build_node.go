package imagebuildserviceimpl

import (
	"context"
	"strings"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

//nolint:gocognit
func (s *service) SelectBuildWorkerNode(
	ctx context.Context,
	buildSetting *entity.ImageBuildSettings,
) (*swarm.Node, error) {
	if buildSetting == nil || (len(buildSetting.Workers.NodeIDs) == 0 && len(buildSetting.Workers.NodeLabels) == 0) {
		return nil, nil
	}

	nodesResp, err := s.dockerManager.NodeList(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	for _, node := range nodesResp.Items {
		if node.Status.State != swarm.NodeStateReady {
			continue
		}
		// Match by NodeIDs
		if len(buildSetting.Workers.NodeIDs) > 0 {
			if gofn.Contain(buildSetting.Workers.NodeIDs, node.ID) ||
				gofn.Contain(buildSetting.Workers.NodeIDs, node.Description.Hostname) {
				return &node, nil
			}
		}
		// Match by NodeLabels
		if len(buildSetting.Workers.NodeLabels) > 0 {
			matched := true
			for _, label := range buildSetting.Workers.NodeLabels {
				parts := strings.SplitN(label, "=", 2) //nolint:mnd
				key := strings.TrimSpace(parts[0])
				val := ""
				if len(parts) == 2 { //nolint:mnd
					val = strings.TrimSpace(parts[1])
				}
				if actualVal, ok := node.Spec.Labels[key]; !ok || (val != "" && actualVal != val) {
					matched = false
					break
				}
			}
			if matched {
				return &node, nil
			}
		}
	}

	return nil, nil
}

package appplacementsettingsuc

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

// checkANodeMatches refuses placement settings that no node in the cluster
// satisfies.
//
// Swarm's answer to a set of constraints nothing satisfies is to leave every
// task Pending, forever, with no error in any place an operator looks. Turning
// that into a refusal at the moment the rule is written is the whole point of
// this check.
func (uc *UC) checkANodeMatches(ctx context.Context, cfg *entity.AppPlacementSettings) error {
	if cfg == nil || (len(cfg.RequireNodeLabels) == 0 && len(cfg.ExcludeNodeLabels) == 0 && !cfg.ExcludeManagerNodes) {
		// Nothing narrows placement, so nothing can rule every node out.
		return nil
	}

	// On a single-node cluster these rules are never applied - loadPlacementSettingsData
	// drops them, which is what the settings page promises - so refusing to save
	// them would refuse a rule that does nothing.
	isMultiNode, err := uc.ClusterService.IsMultiNode(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if !isMultiNode {
		return nil
	}

	resp, err := uc.dockerManager.NodeList(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if len(resp.Items) == 0 {
		// No cluster to judge against: saying "no node matches" here would be a
		// guess dressed as a fact.
		return nil
	}
	if hasMatchingNode(resp.Items, cfg) {
		return nil
	}
	return hperrors.Wrap(hperrors.ErrPlacementNoMatchingNode)
}

// hasMatchingNode reports whether any node satisfies the settings.
//
// It judges labels and role only. The build-node exclusion depends on the image
// build settings and a volume's node pin is per app, so neither is known here -
// which means this check catches the rule that excludes everything, not every
// way an app can end up unschedulable. Node health is ignored on purpose: a
// machine that is rebooting still satisfies a label rule, and refusing to save
// settings because of it would be worse than the problem.
func hasMatchingNode(nodes []swarm.Node, cfg *entity.AppPlacementSettings) bool {
	for i := range nodes {
		if nodeMatches(&nodes[i], cfg) {
			return true
		}
	}
	return false
}

func nodeMatches(node *swarm.Node, cfg *entity.AppPlacementSettings) bool {
	if cfg.ExcludeManagerNodes && node.Spec.Role == swarm.NodeRoleManager {
		return false
	}
	for _, selector := range cfg.RequireNodeLabels {
		key, value, ok := dockerhelper.ParseNodeLabelSelector(selector)
		// A selector that cannot be parsed was refused when the settings were
		// saved; here it is simply not a rule.
		if ok && node.Spec.Labels[key] != value {
			return false
		}
	}
	for _, selector := range cfg.ExcludeNodeLabels {
		key, value, ok := dockerhelper.ParseNodeLabelSelector(selector)
		if ok && node.Spec.Labels[key] == value {
			return false
		}
	}
	return true
}

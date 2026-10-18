package artifactserviceimpl

import (
	"context"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

func (s *service) ArtifactsOf(ctx context.Context, db database.IDB, appID string) (
	*entity.AppArtifactsSettings, error) {
	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppArtifacts),
		bunex.SelectWhere("setting.object_id = ?", appID),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(settings) == 0 {
		return nil, nil
	}
	artifacts, err := settings[0].AsAppArtifactsSettings()
	return artifacts, hperrors.Wrap(err)
}

func (s *service) MountsFor(
	ctx context.Context,
	artifacts *entity.AppArtifactsSettings,
	spec *swarm.ServiceSpec,
) ([]mount.Mount, error) {
	if artifacts == nil || len(artifacts.Artifacts) == 0 {
		return nil, nil
	}
	nodes, err := s.eligibleNodes(ctx, spec)
	if err != nil {
		return nil, err
	}
	mounts := make([]mount.Mount, 0, len(artifacts.Artifacts))
	for _, artifact := range artifacts.Artifacts {
		if artifact == nil {
			continue
		}
		if err = sameBuildEverywhere(artifact, nodes); err != nil {
			return nil, err
		}
		var hostPath string
		for _, node := range nodes {
			ensured, ensureErr := s.EnsureOnNode(ctx, &artifactservice.EnsureOnNodeReq{
				NodeID: node.ID, Artifact: artifact,
			})
			if ensureErr != nil {
				return nil, ensureErr
			}
			hostPath = ensured.HostPath
		}
		mounts = append(mounts, artifactservice.Mount(hostPath, artifact.MountPath))
	}
	return mounts, nil
}

// sameBuildEverywhere refuses nodes that would need different files: one bind
// mount names one source for every task of the service.
func sameBuildEverywhere(artifact *entity.AppArtifact, nodes []swarm.Node) error {
	var first *swarm.Node
	var firstSHA string
	for i := range nodes {
		node := &nodes[i]
		arch, _ := base.CPUArchFromDocker(node.Description.Platform.Architecture)
		sha := ""
		if file := artifact.Files[arch]; file != nil {
			sha = file.SHA256
		}
		if first == nil {
			first, firstSHA = node, sha
			continue
		}
		if sha != firstSHA {
			return hperrors.Wrap(hperrors.ErrArtifactArchUnsupported).
				WithParam("Name", artifact.Name).WithParam("Node", node.Description.Hostname).
				WithParam("Arch", node.Description.Platform.Architecture).
				WithExtraDetail("the app may run on %s (%s) as well, and one mount cannot name two builds: "+
					"pin the app to one node", first.Description.Hostname, first.Description.Platform.Architecture)
		}
	}
	return nil
}

// eligibleNodes are the nodes a task of spec could start on: Ready, active, and
// matching every placement constraint. A constraint on something this does not
// read - engine labels - is taken as met, so it can only widen the set.
func (s *service) eligibleNodes(ctx context.Context, spec *swarm.ServiceSpec) ([]swarm.Node, error) {
	listed, err := s.dockerManager.NodeList(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	var constraints []string
	if spec.TaskTemplate.Placement != nil {
		constraints = spec.TaskTemplate.Placement.Constraints
	}
	var nodes []swarm.Node
	for _, node := range listed.Items {
		if node.Status.State != swarm.NodeStateReady || node.Spec.Availability != swarm.NodeAvailabilityActive {
			continue
		}
		if satisfiesAll(&node, constraints) {
			nodes = append(nodes, node)
		}
	}
	if len(nodes) == 0 {
		return nil, hperrors.Wrap(hperrors.ErrArtifactDownloadFailed).WithParam("Name", "artifacts").
			WithExtraDetail("no Ready node satisfies the app's placement (%s)", strings.Join(constraints, ", "))
	}
	return nodes, nil
}

func satisfiesAll(node *swarm.Node, constraints []string) bool {
	for _, constraint := range constraints {
		key, op, want := dockerhelper.ParsePlacementConstraint(constraint)
		have, known := nodeValue(node, key)
		if key == "" || !known {
			continue
		}
		if equal := strings.EqualFold(have, want); equal != (op == "==") {
			return false
		}
	}
	return true
}

// nodeValue is what a placement constraint's key reads of a node, as Swarm
// reads it.
func nodeValue(node *swarm.Node, key string) (string, bool) {
	switch key {
	case "node.id":
		return node.ID, true
	case "node.hostname":
		return node.Description.Hostname, true
	case "node.role":
		return string(node.Spec.Role), true
	case "node.platform.os":
		return node.Description.Platform.OS, true
	case "node.platform.arch":
		return node.Description.Platform.Architecture, true
	}
	if label, found := strings.CutPrefix(key, "node.labels."); found {
		return node.Spec.Labels[label], true
	}
	return "", false
}

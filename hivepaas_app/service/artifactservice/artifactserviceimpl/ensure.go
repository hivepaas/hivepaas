package artifactserviceimpl

import (
	"context"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/artifactagentuc/artifactagentdto"
)

func (s *service) EnsureOnNode(
	ctx context.Context,
	req *artifactservice.EnsureOnNodeReq,
) (*artifactservice.EnsureOnNodeResp, error) {
	inspect, err := s.dockerManager.NodeInspect(ctx, req.NodeID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	description := inspect.Node.Description
	arch, known := base.CPUArchFromDocker(description.Platform.Architecture)
	file := req.Artifact.Files[arch]
	if !known || file == nil {
		reported := string(arch)
		if !known {
			reported = description.Platform.Architecture
		}
		node := description.Hostname
		if node == "" {
			node = req.NodeID
		}
		return nil, hperrors.Wrap(hperrors.ErrArtifactArchUnsupported).
			WithParam("Name", req.Artifact.Name).WithParam("Node", node).WithParam("Arch", reported)
	}

	agentAddr, err := s.agentService.GetAgentAddrForNode(ctx, req.NodeID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	agent, err := s.newClient(agentAddr)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	defer func() { _ = agent.Close() }()

	ensured, err := agent.Ensure(ctx, &artifactagentdto.EnsureReq{
		URL:       file.URL,
		MirrorURL: s.mirrorURL(file.SHA256),
		SHA256:    file.SHA256,
		Size:      file.Size,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &artifactservice.EnsureOnNodeResp{
		HostPath:   ensured.HostPath,
		Arch:       arch,
		SHA256:     file.SHA256,
		Downloaded: ensured.Downloaded,
	}, nil
}

// mirrorURL is where the mirror has a file: under its base URL, by hash. Empty
// when no mirror is configured.
func (s *service) mirrorURL(sha string) string {
	base := strings.TrimSuffix(s.mirror(), "/")
	if base == "" {
		return ""
	}
	return base + "/" + sha
}

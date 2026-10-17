package artifactservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/artifactagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/artifactagentuc/artifactagentdto"
)

func Ensure(
	ctx context.Context,
	uc *artifactagentuc.UC,
	req *agentproto.EnsureArtifactReq,
) (*agentproto.EnsureArtifactResp, error) {
	resp, err := uc.Ensure(ctx, &artifactagentdto.EnsureReq{
		URL:       req.GetUrl(),
		MirrorURL: req.GetMirrorUrl(),
		SHA256:    req.GetSha256(),
		Size:      req.GetSize(),
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &agentproto.EnsureArtifactResp{HostPath: resp.HostPath, Downloaded: resp.Downloaded}, nil
}

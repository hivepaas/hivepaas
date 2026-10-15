// Package imagescanservice serves the agent's image scan calls: a scan on its
// node.
package imagescanservice

import (
	"context"
	"encoding/json"
	"time"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice/imagescanrun"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagescanagentuc"
)

// ImageScanRun runs the scan the call carries, on this node.
func ImageScanRun(
	ctx context.Context,
	uc *imagescanagentuc.UC,
	req *agentproto.ImageScanRunReq,
) (*agentproto.ImageScanRunResp, error) {
	var volume mount.Mount
	if err := json.Unmarshal(req.GetVolumeMount(), &volume); err != nil {
		return nil, hperrors.ToGRPCError(hperrors.Wrap(err)) //nolint:wrapcheck
	}
	resp, err := uc.Run(ctx, &imagescanrun.RunReq{
		TrivyImage:       req.GetTrivyImage(),
		ImageID:          req.GetImageId(),
		ImageNode:        req.GetImageNode(),
		RemoteRef:        req.GetRemoteRef(),
		Platform:         req.GetPlatform(),
		DockerConfig:     req.GetDockerConfig(),
		Volume:           volume,
		CacheDir:         req.GetCacheDir(),
		DBRepository:     req.GetDbRepository(),
		JavaDBRepository: req.GetJavaDbRepository(),
		NanoCPUs:         req.GetNanoCpus(),
		MemoryBytes:      req.GetMemoryBytes(),
		Timeout:          time.Duration(req.GetTimeoutSeconds()) * time.Second,
		Format:           req.GetFormat(),
	})
	if err != nil {
		return nil, hperrors.ToGRPCError(err) //nolint:wrapcheck
	}
	return &agentproto.ImageScanRunResp{
		Source:       string(resp.Source),
		Output:       resp.Output,
		Log:          resp.Log,
		LogTruncated: resp.LogTruncated,
		OomKilled:    resp.OOMKilled,
		ExitCode:     resp.ExitCode,
	}, nil
}

package imagebuildservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc/imagebuildagentdto"
)

func ImageBuild(
	uc *imagebuildagentuc.UC,
	req *agentproto.ImageBuildReq,
	stream agentproto.ImageBuildService_ImageBuildServer,
) error {
	if req == nil {
		return nil
	}

	var dockerfile entity.DeploymentDockerfile
	if df := req.GetDockerfile(); df != nil {
		dockerfile = entity.DeploymentDockerfile{
			Source:   base.DockerfileSource(df.GetSource()),
			Path:     df.GetPath(),
			Content:  df.GetContent(),
			ScanPath: df.GetScanPath(),
		}
	}

	var buildSettings *entity.ImageBuildSettings
	//nolint:gosec
	if bs := req.GetImageBuildSettings(); bs != nil {
		buildSettings = &entity.ImageBuildSettings{
			NoCache:   bs.GetNoCache(),
			NoVerbose: bs.GetNoVerbose(),
		}
		if bs.GetWorkers() != nil {
			buildSettings.Workers = entity.ImageBuildWorkerSettings{
				NodeIDs:    bs.GetWorkers().GetNodeIds(),
				NodeLabels: bs.GetWorkers().GetNodeLabels(),
			}
		}
		if bs.GetResources() != nil {
			buildSettings.Resources = entity.ImageBuildResourceSettings{
				CPUs:    uint(bs.GetResources().GetCpus()),
				Mem:     unit.DataSize(bs.GetResources().GetMem()),
				MemSwap: unit.DataSize(bs.GetResources().GetMemSwap()),
				ShmSize: unit.DataSize(bs.GetResources().GetShmSize()),
			}
		}
		if bs.GetSources() != nil {
			buildSettings.Sources = entity.ImageBuildSourceSettings{
				RepoCache: bs.GetSources().GetRepoCache(),
			}
		}
	}

	dtoReq := &imagebuildagentdto.ImageBuildReq{
		TaskID: req.GetTaskId(),
		AppID:  req.GetAppId(),
		ImageBuildReq: imagebuildservice.ImageBuildReq{
			CommitHash:         req.GetCommitHash(),
			Dockerfile:         dockerfile,
			ImageName:          req.GetImageName(),
			PushToRegistry:     entity.ObjectID{ID: req.GetPushToRegistryId()},
			ImageBuildSettings: buildSettings,
			NoCache:            req.GetNoCache(),
			BuildID:            req.GetBuildId(),
			CheckoutDir:        req.GetCheckoutDir(),
			TempDir:            req.GetTempDir(),
		},
		SendLog: func(frames []*tasklog.LogFrame) error {
			for _, frame := range frames {
				resp := &agentproto.ImageBuildResp{
					Value: &agentproto.ImageBuildResp_Log{
						Log: &agentproto.LogFrame{
							Type: string(frame.Type),
							Data: frame.Data,
							Ts:   frame.Ts.UnixNano(),
						},
					},
				}
				if err := stream.Send(resp); err != nil {
					return apperrors.Wrap(err)
				}
			}
			return nil
		},
	}

	resp, err := uc.ImageBuild(stream.Context(), dtoReq)
	if err != nil {
		return apperrors.ToGRPCError(err) //nolint:wrapcheck
	}

	if resp != nil {
		resultResp := &agentproto.ImageBuildResp{
			Value: &agentproto.ImageBuildResp_Result{
				Result: &agentproto.ImageBuildResult{
					ImageTags: resp.ImageTags,
				},
			},
		}
		if err := stream.Send(resultResp); err != nil {
			return apperrors.ToGRPCError(err) //nolint:wrapcheck
		}
	}

	return nil
}

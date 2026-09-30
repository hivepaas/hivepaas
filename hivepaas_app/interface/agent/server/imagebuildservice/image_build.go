package imagebuildservice

import (
	"github.com/moby/moby/api/types/registry"
	"google.golang.org/grpc"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc/imagebuildagentdto"
)

// ImageBuildFromSource builds an image from the source the call carries: the
// request in its first message, the packed source in those that follow.
func ImageBuildFromSource(
	uc *imagebuildagentuc.UC,
	stream grpc.BidiStreamingServer[agentproto.ImageBuildMsg, agentproto.ImageBuildResp],
) error {
	first, err := stream.Recv()
	if err != nil {
		return hperrors.ToGRPCError(err) //nolint:wrapcheck
	}
	req := first.GetReq()
	if req == nil {
		return hperrors.ToGRPCError(hperrors.NewMissing("Image build request")) //nolint:wrapcheck
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
				NodeIDs:        bs.GetWorkers().GetNodeIds(),
				NodeLabels:     bs.GetWorkers().GetNodeLabels(),
				MaxParallelism: int(bs.GetWorkers().GetMaxParallelism()),
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
			ImageTags:          req.GetImageTags(),
			PushToRegistry:     entity.ObjectID{ID: req.GetPushToRegistryId()},
			ImageBuildSettings: buildSettings,
			NoCache:            req.GetNoCache(),
			BuildID:            req.GetBuildId(),
			Inputs:             inputsFromProto(req.GetInputs()),
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
					return hperrors.Wrap(err)
				}
			}
			return nil
		},
	}

	resp, err := uc.ImageBuildFromSource(stream.Context(), dtoReq, &sourceReader{stream: stream})
	if err != nil {
		return hperrors.ToGRPCError(err) //nolint:wrapcheck
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
			return hperrors.ToGRPCError(err) //nolint:wrapcheck
		}
	}

	return nil
}

// inputsFromProto is what the app resolved for the build. A request without it
// stays without it, and the agent refuses the build.
func inputsFromProto(in *agentproto.ImageBuildInputs) *imagebuildservice.BuildInputs {
	if in == nil {
		return nil
	}
	out := &imagebuildservice.BuildInputs{
		EnvVars:       make(map[string]*string, len(in.GetEnvVars())),
		RegistryAuths: make(map[string]registry.AuthConfig, len(in.GetRegistryAuths())),
		Secrets:       in.GetSecrets(),
	}
	for key, value := range in.GetEnvVars() {
		out.EnvVars[key] = &value
	}
	for _, auth := range in.GetRegistryAuths() {
		out.RegistryAuths[auth.GetAddress()] = registryAuthFromProto(auth)
	}
	if push := in.GetPushRegistry(); push != nil {
		auth := registryAuthFromProto(push)
		out.PushRegistry = &auth
	}
	return out
}

func registryAuthFromProto(auth *agentproto.ImageBuildRegistryAuth) registry.AuthConfig {
	return registry.AuthConfig{
		Username:      auth.GetUsername(),
		Password:      auth.GetPassword(),
		ServerAddress: auth.GetAddress(),
	}
}

// sourceReader is the packed source, read from the chunks that follow the request.
type sourceReader struct {
	stream grpc.BidiStreamingServer[agentproto.ImageBuildMsg, agentproto.ImageBuildResp]
	chunk  []byte
}

func (r *sourceReader) Read(p []byte) (int, error) {
	for len(r.chunk) == 0 {
		msg, err := r.stream.Recv()
		if err != nil {
			return 0, err //nolint:wrapcheck // io.EOF ends the source
		}
		r.chunk = msg.GetSourceChunk()
	}
	n := copy(p, r.chunk)
	r.chunk = r.chunk[n:]
	return n, nil
}

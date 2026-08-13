package imagebuildservice

import (
	"context"
	"errors"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc/imagebuildagentdto"
)

type ImageBuildServiceClient interface {
	ImageBuild(ctx context.Context, req *imagebuildagentdto.ImageBuildReq) (*imagebuildagentdto.ImageBuildResp, error)
	Close() error
}

type grpcImageBuildServiceClient struct {
	protoClient agentproto.ImageBuildServiceClient
	conn        *grpc.ClientConn
}

func NewImageBuildServiceClient(agentAddr string) (ImageBuildServiceClient, error) {
	conn, err := grpc.NewClient(agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return &grpcImageBuildServiceClient{
		conn:        conn,
		protoClient: agentproto.NewImageBuildServiceClient(conn),
	}, nil
}

func (c *grpcImageBuildServiceClient) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}

func (c *grpcImageBuildServiceClient) ImageBuild(
	ctx context.Context,
	req *imagebuildagentdto.ImageBuildReq,
) (*imagebuildagentdto.ImageBuildResp, error) {
	authCtx := client.CreateAuthCtx(ctx)

	var protoDockerfile *agentproto.DeploymentDockerfile
	if req.Dockerfile.Source != "" || req.Dockerfile.Path != "" || req.Dockerfile.Content != "" ||
		req.Dockerfile.ScanPath != "" {
		protoDockerfile = &agentproto.DeploymentDockerfile{
			Source:   string(req.Dockerfile.Source),
			Path:     req.Dockerfile.Path,
			Content:  req.Dockerfile.Content,
			ScanPath: req.Dockerfile.ScanPath,
		}
	}

	var protoBuildSettings *agentproto.ImageBuildSettings
	//nolint:gosec
	if req.ImageBuildSettings != nil {
		protoBuildSettings = &agentproto.ImageBuildSettings{
			NoCache:   req.ImageBuildSettings.NoCache,
			NoVerbose: req.ImageBuildSettings.NoVerbose,
			Workers: &agentproto.ImageBuildWorkerSettings{
				NodeIds:    req.ImageBuildSettings.Workers.NodeIDs,
				NodeLabels: req.ImageBuildSettings.Workers.NodeLabels,
			},
			Resources: &agentproto.ImageBuildResourceSettings{
				Cpus:    uint32(req.ImageBuildSettings.Resources.CPUs),
				Mem:     uint64(req.ImageBuildSettings.Resources.Mem),
				MemSwap: uint64(req.ImageBuildSettings.Resources.MemSwap),
				ShmSize: uint64(req.ImageBuildSettings.Resources.ShmSize),
			},
			Sources: &agentproto.ImageBuildSourceSettings{
				RepoCache: req.ImageBuildSettings.Sources.RepoCache,
			},
		}
	}

	appID := req.AppID
	if appID == "" && req.App != nil {
		appID = req.App.ID
	}

	protoReq := &agentproto.ImageBuildReq{
		TaskId:             req.TaskID,
		AppId:              appID,
		CommitHash:         req.CommitHash,
		Dockerfile:         protoDockerfile,
		ImageName:          req.ImageName,
		PushToRegistryId:   req.PushToRegistry.ID,
		ImageBuildSettings: protoBuildSettings,
		NoCache:            req.NoCache,
		BuildId:            req.BuildID,
		CheckoutDir:        req.CheckoutDir,
		TempDir:            req.TempDir,
	}

	stream, err := c.protoClient.ImageBuild(authCtx, protoReq)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respDTO := &imagebuildagentdto.ImageBuildResp{}

	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, apperrors.Wrap(err)
		}

		if logFrame := resp.GetLog(); logFrame != nil {
			frame := &tasklog.LogFrame{
				Type: tasklog.LogType(logFrame.GetType()),
				Data: logFrame.GetData(),
				Ts:   time.Unix(0, logFrame.GetTs()),
			}
			if req.SendLog != nil {
				_ = req.SendLog([]*tasklog.LogFrame{frame})
			} else if req.LogStore != nil {
				_ = req.LogStore.Add(ctx, frame)
			}
		}

		if result := resp.GetResult(); result != nil {
			respDTO.ImageTags = result.GetImageTags()
		}
	}

	return respDTO, nil
}

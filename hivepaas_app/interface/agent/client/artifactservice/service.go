package artifactservice

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/artifactagentuc/artifactagentdto"
)

type ArtifactServiceClient interface {
	Ensure(ctx context.Context, req *artifactagentdto.EnsureReq) (*artifactagentdto.EnsureResp, error)
	Close() error
}

type grpcArtifactServiceClient struct {
	protoClient agentproto.ArtifactServiceClient
	conn        *grpc.ClientConn
}

func NewArtifactServiceClient(agentAddr string) (ArtifactServiceClient, error) {
	conn, err := grpc.NewClient(agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &grpcArtifactServiceClient{
		conn:        conn,
		protoClient: agentproto.NewArtifactServiceClient(conn),
	}, nil
}

func (c *grpcArtifactServiceClient) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (c *grpcArtifactServiceClient) Ensure(
	ctx context.Context,
	req *artifactagentdto.EnsureReq,
) (*artifactagentdto.EnsureResp, error) {
	resp, err := c.protoClient.Ensure(client.CreateAuthCtx(ctx), &agentproto.EnsureArtifactReq{
		Url:       req.URL,
		MirrorUrl: req.MirrorURL,
		Sha256:    req.SHA256,
		Size:      req.Size,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &artifactagentdto.EnsureResp{HostPath: resp.GetHostPath(), Downloaded: resp.GetDownloaded()}, nil
}

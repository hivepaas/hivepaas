package volumeservice

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/volumeagentuc/volumeagentdto"
)

type VolumeServiceClient interface {
	RemoveVolume(ctx context.Context, req *volumeagentdto.RemoveVolumeReq) (*volumeagentdto.RemoveVolumeResp, error)
	Close() error
}

type grpcVolumeServiceClient struct {
	protoClient agentproto.VolumeServiceClient
	conn        *grpc.ClientConn
}

func NewVolumeServiceClient(agentAddr string) (VolumeServiceClient, error) {
	conn, err := grpc.NewClient(agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &grpcVolumeServiceClient{
		conn:        conn,
		protoClient: agentproto.NewVolumeServiceClient(conn),
	}, nil
}

func (c *grpcVolumeServiceClient) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (c *grpcVolumeServiceClient) RemoveVolume(
	ctx context.Context,
	req *volumeagentdto.RemoveVolumeReq,
) (*volumeagentdto.RemoveVolumeResp, error) {
	authCtx := client.CreateAuthCtx(ctx)

	_, err := c.protoClient.RemoveVolume(authCtx, &agentproto.RemoveVolumeReq{
		VolumeId: req.VolumeID,
		Force:    req.Force,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &volumeagentdto.RemoveVolumeResp{}, nil
}

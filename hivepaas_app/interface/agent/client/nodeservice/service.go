package nodeservice

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodeagentuc/nodeagentdto"
)

type NodeServiceClient interface {
	ExecuteCommand(ctx context.Context, req *nodeagentdto.ExecCommandReq) (*nodeagentdto.ExecCommandResp, error)
	Close() error
}

type grpcNodeServiceClient struct {
	protoClient agentproto.NodeServiceClient
	conn        *grpc.ClientConn
}

func NewNodeServiceClient(agentAddr string) (NodeServiceClient, error) {
	conn, err := grpc.NewClient(agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return &grpcNodeServiceClient{
		conn:        conn,
		protoClient: agentproto.NewNodeServiceClient(conn),
	}, nil
}

func (c *grpcNodeServiceClient) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}

func (c *grpcNodeServiceClient) ExecuteCommand(
	ctx context.Context,
	req *nodeagentdto.ExecCommandReq,
) (*nodeagentdto.ExecCommandResp, error) {
	authCtx := client.CreateAuthCtx(ctx)

	protoReq := &agentproto.ExecCommandReq{
		Command:    req.Command,
		Env:        req.Env,
		WorkingDir: req.WorkingDir,
	}

	resp, err := c.protoClient.ExecuteCommand(authCtx, protoReq)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if req.Stdout != nil && len(resp.GetStdout()) > 0 {
		_, _ = req.Stdout.Write(resp.GetStdout())
	}
	if req.Stderr != nil && len(resp.GetStderr()) > 0 {
		_, _ = req.Stderr.Write(resp.GetStderr())
	} else if req.Stdout != nil && len(resp.GetStderr()) > 0 {
		_, _ = req.Stdout.Write(resp.GetStderr())
	}

	return &nodeagentdto.ExecCommandResp{
		ExitCode: resp.GetExitCode(),
	}, nil
}

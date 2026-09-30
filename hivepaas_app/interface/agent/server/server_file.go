package server

import (
	"context"

	"google.golang.org/grpc"

	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server/fileservice"
)

func (s *AgentServer) FileRead(
	req *agentproto.FileReq,
	stream grpc.ServerStreamingServer[agentproto.FileChunk],
) error {
	return fileservice.FileRead(s.fileAgentUC, req, stream) //nolint:wrapcheck
}

func (s *AgentServer) FileWrite(
	stream grpc.ClientStreamingServer[agentproto.FileWriteReq, agentproto.FileWriteResp],
) error {
	return fileservice.FileWrite(s.fileAgentUC, stream) //nolint:wrapcheck
}

func (s *AgentServer) FileRemove(
	_ context.Context,
	req *agentproto.FileReq,
) (*agentproto.FileRemoveResp, error) {
	return fileservice.FileRemove(s.fileAgentUC, req) //nolint:wrapcheck
}

func (s *AgentServer) FileStat(
	_ context.Context,
	req *agentproto.FileReq,
) (*agentproto.FileStatResp, error) {
	return fileservice.FileStat(s.fileAgentUC, req) //nolint:wrapcheck
}

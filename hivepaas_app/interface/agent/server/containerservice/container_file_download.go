package containerservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc/containeragentdto"
)

type grpcDownloadStream struct {
	stream agentproto.ContainerService_ContainerCopyFromServer
}

func (s *grpcDownloadStream) Send(out *containeragentdto.DownloadFileOutput) error {
	return s.stream.Send(&agentproto.ContainerCopyFromResp{ //nolint:wrapcheck
		Chunk:       out.Chunk,
		FileName:    out.FileName,
		FileSize:    out.FileSize,
		ContentType: out.ContentType,
	})
}

// ContainerCopyFrom extracts a file or directory archive from a container.
func ContainerCopyFrom(
	containerAgentUC *containeragentuc.UC,
	req *agentproto.ContainerCopyFromReq,
	stream agentproto.ContainerService_ContainerCopyFromServer,
) error {
	wrappedStream := &grpcDownloadStream{stream: stream}
	err := containerAgentUC.DownloadFile(
		stream.Context(),
		&containeragentdto.DownloadFileInput{
			ContainerID:       req.GetContainerId(),
			Path:              req.GetPath(),
			IsDir:             req.GetIsDir(),
			CompressionFormat: base.FileCompressionFormat(req.GetCompressionFormat()),
		},
		wrappedStream,
	)
	return apperrors.ToGRPCError(err) //nolint:wrapcheck
}

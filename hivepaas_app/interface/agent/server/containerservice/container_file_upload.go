package containerservice

import (
	"errors"
	"io"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc/containeragentdto"
)

// ContainerCopyTo receives a TAR stream from client and copies it into container.
func ContainerCopyTo(
	containerAgentUC *containeragentuc.UC,
	stream agentproto.ContainerService_ContainerCopyToServer,
) error {
	firstReq, err := stream.Recv()
	if err != nil {
		return hperrors.ToGRPCError(hperrors.Wrap(err)) //nolint:wrapcheck
	}

	metadata := firstReq.GetMetadata()
	if metadata == nil {
		//nolint:wrapcheck
		return hperrors.ToGRPCError(hperrors.NewArgumentInvalid("metadata").
			WithMsgLog("first message must contain metadata"))
	}

	pr, pw := io.Pipe()

	go func() {
		for {
			req, recvErr := stream.Recv()
			if recvErr != nil {
				if errors.Is(recvErr, io.EOF) {
					_ = pw.Close()
					return
				}
				_ = pw.CloseWithError(recvErr)
				return
			}

			if chunk := req.GetChunk(); len(chunk) > 0 {
				if _, writeErr := pw.Write(chunk); writeErr != nil {
					_ = pw.CloseWithError(writeErr)
					return
				}
			}
		}
	}()

	uploadErr := containerAgentUC.UploadFile(
		stream.Context(),
		&containeragentdto.UploadFileInput{
			ContainerID: metadata.GetContainerId(),
			DstPath:     metadata.GetDstPath(),
			TarReader:   pr,
			Overwrite:   metadata.GetAllowOverwriteDirWithFile(),
			CopyUIDGID:  metadata.GetCopyUidGid(),
		},
	)
	if uploadErr != nil {
		_ = pr.Close()
		return hperrors.ToGRPCError(uploadErr) //nolint:wrapcheck
	}

	return stream.SendAndClose(&agentproto.ContainerCopyToResp{ //nolint:wrapcheck
		Success: true,
		Message: "File uploaded successfully",
	})
}

package containerservice

import (
	"context"
	"errors"
	"io"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc/containeragentdto"
)

const (
	uploadChunkSize = 32 * 1024 // 32KB
)

// ContainerCopyTo streams a TAR archive into a remote container.
func (c *grpcContainerServiceClient) ContainerCopyTo(
	ctx context.Context,
	req *containeragentdto.UploadFileInput,
) error {
	authCtx, cancelFunc := client.CreateAuthCtxWithCancel(ctx)
	defer cancelFunc()

	stream, err := c.protoClient.ContainerCopyTo(authCtx)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// 1. Send metadata
	if err := stream.Send(&agentproto.ContainerCopyToReq{
		Value: &agentproto.ContainerCopyToReq_Metadata{
			Metadata: &agentproto.ContainerCopyToMetadata{
				ContainerId:               req.ContainerID,
				DstPath:                   req.DstPath,
				AllowOverwriteDirWithFile: req.Overwrite,
				CopyUidGid:                req.CopyUIDGID,
			},
		},
	}); err != nil {
		return hperrors.Wrap(err)
	}

	// 2. Stream chunks
	buf := make([]byte, uploadChunkSize)
	for {
		n, readErr := req.TarReader.Read(buf)
		if n > 0 {
			if err := stream.Send(&agentproto.ContainerCopyToReq{
				Value: &agentproto.ContainerCopyToReq_Chunk{
					Chunk: buf[:n],
				},
			}); err != nil {
				return hperrors.Wrap(err)
			}
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return hperrors.Wrap(readErr)
		}
	}

	// 3. Close and receive response
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return hperrors.Wrap(err)
	}

	if !resp.GetSuccess() {
		return hperrors.NewInternal().WithMsgLog("failed to copy to container: %s", resp.GetMessage())
	}

	return nil
}

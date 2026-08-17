package containerservice

import (
	"context"
	"errors"
	"io"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc/containeragentdto"
)

type RemoteFileDownloadResult struct {
	FileName    string
	FileSize    int64
	ContentType string
	Reader      io.ReadCloser
}

type grpcDownloadReader struct {
	pr         *io.PipeReader
	cancelFunc context.CancelFunc
}

func (r *grpcDownloadReader) Read(p []byte) (int, error) {
	n, err := r.pr.Read(p)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return n, io.EOF
		}
		return n, apperrors.Wrap(err)
	}
	return n, nil
}

func (r *grpcDownloadReader) Close() error {
	if r.cancelFunc != nil {
		r.cancelFunc()
	}
	if err := r.pr.Close(); err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

// ContainerCopyFrom starts a file download stream from remote container.
func (c *grpcContainerServiceClient) ContainerCopyFrom(
	ctx context.Context,
	req *containeragentdto.DownloadFileInput,
) (*RemoteFileDownloadResult, error) {
	authCtx, cancelFunc := client.CreateAuthCtxWithCancel(ctx)
	stream, err := c.protoClient.ContainerCopyFrom(authCtx, &agentproto.ContainerCopyFromReq{
		ContainerId:       req.ContainerID,
		Path:              req.Path,
		IsDir:             req.IsDir,
		CompressionFormat: string(req.CompressionFormat),
	})

	if err != nil {
		cancelFunc()
		return nil, apperrors.Wrap(err)
	}

	firstResp, err := stream.Recv()
	if err != nil {
		cancelFunc()
		return nil, apperrors.Wrap(err)
	}

	pr, pw := io.Pipe()

	go func() {
		defer cancelFunc()

		if chunk := firstResp.GetChunk(); len(chunk) > 0 {
			if _, writeErr := pw.Write(chunk); writeErr != nil {
				_ = pw.CloseWithError(apperrors.Wrap(writeErr))
				return
			}
		}

		for {
			resp, recvErr := stream.Recv()
			if recvErr != nil {
				if errors.Is(recvErr, io.EOF) {
					_ = pw.Close()
					return
				}
				_ = pw.CloseWithError(apperrors.Wrap(recvErr))
				return
			}

			if chunk := resp.GetChunk(); len(chunk) > 0 {
				if _, writeErr := pw.Write(chunk); writeErr != nil {
					_ = pw.CloseWithError(apperrors.Wrap(writeErr))
					return
				}
			}
		}
	}()

	return &RemoteFileDownloadResult{
		FileName:    firstResp.GetFileName(),
		FileSize:    firstResp.GetFileSize(),
		ContentType: firstResp.GetContentType(),
		Reader: &grpcDownloadReader{
			pr:         pr,
			cancelFunc: cancelFunc,
		},
	}, nil
}

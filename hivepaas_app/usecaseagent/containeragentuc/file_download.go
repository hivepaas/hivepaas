package containeragentuc

import (
	"context"
	"errors"
	"io"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerfileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc/containeragentdto"
)

const (
	downloadChunkSize = 32 * 1024 // 32KB
)

func (uc *UC) DownloadFile(
	ctx context.Context,
	input *containeragentdto.DownloadFileInput,
	stream containeragentdto.DownloadFileStream,
) error {
	res, err := uc.dockerManager.ContainerCopyFrom(ctx, input.ContainerID, input.Path)
	if err != nil {
		return hperrors.Wrap(err)
	}

	resp, err := uc.containerFileService.PrepareDownloadStream(ctx, &containerfileservice.PrepareDownloadStreamReq{
		Path:              input.Path,
		IsDir:             input.IsDir,
		CompressionFormat: input.CompressionFormat,
		Content:           res.Content,
		Stat:              &res.Stat,
	})
	if err != nil {
		_ = res.Content.Close()
		return hperrors.Wrap(err)
	}
	defer resp.Reader.Close()

	buf := make([]byte, downloadChunkSize)
	isFirst := true

	for {
		n, readErr := resp.Reader.Read(buf)
		if n > 0 {
			out := &containeragentdto.DownloadFileOutput{
				Chunk: buf[:n],
			}
			if isFirst {
				out.FileName = resp.FileName
				out.FileSize = resp.FileSize
				out.ContentType = resp.ContentType
				isFirst = false
			}
			if sendErr := stream.Send(out); sendErr != nil {
				return hperrors.Wrap(sendErr)
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return hperrors.Wrap(readErr)
		}
	}

	return nil
}

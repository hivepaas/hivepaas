package containeragentuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc/containeragentdto"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (uc *UC) UploadFile(
	ctx context.Context,
	input *containeragentdto.UploadFileInput,
) error {
	opts := make([]docker.ContainerCopyToOption, 0, 2) //nolint:mnd
	if input.Overwrite {
		opts = append(opts, docker.ContainerCopyToWithAllowOverwriteDirWithFile(true))
	}
	if input.CopyUIDGID {
		opts = append(opts, docker.ContainerCopyToWithCopyUIDGID(true))
	}

	_, err := uc.dockerManager.ContainerCopyTo(ctx, input.ContainerID, input.DstPath, input.TarReader, opts...)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

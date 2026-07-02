package fileuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc/filedto"
)

func (uc *UC) UpdateFileStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.UpdateFileStatusReq,
) (*filedto.UpdateFileStatusResp, error) {
	// TODO: add implementation
	return nil, nil
}

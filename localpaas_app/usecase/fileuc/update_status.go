package fileuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/fileuc/filedto"
)

func (uc *UC) UpdateFileStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.UpdateFileStatusReq,
) (*filedto.UpdateFileStatusResp, error) {
	// TODO: add implementation
	return nil, nil
}

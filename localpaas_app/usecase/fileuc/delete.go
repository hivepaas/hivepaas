package fileuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/fileuc/filedto"
)

func (uc *UC) DeleteFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.DeleteFileReq,
) (*filedto.DeleteFileResp, error) {
	// TODO: add implementation
	return &filedto.DeleteFileResp{}, nil
}

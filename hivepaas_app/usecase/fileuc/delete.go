package fileuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc/filedto"
)

func (uc *UC) DeleteFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.DeleteFileReq,
) (*filedto.DeleteFileResp, error) {
	// TODO: add implementation
	return &filedto.DeleteFileResp{}, nil
}

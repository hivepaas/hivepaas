package containerfileservice

import (
	"context"
)

type Service interface {
	PrepareDownloadStream(ctx context.Context, req *PrepareDownloadStreamReq) (*PrepareDownloadStreamResp, error)
	PrepareUploadTarStream(ctx context.Context, req *PrepareUploadTarStreamReq) (*PrepareUploadTarStreamResp, error)
}

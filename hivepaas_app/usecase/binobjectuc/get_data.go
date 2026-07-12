package binobjectuc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/binobjectuc/binobjectdto"
)

const (
	defaultImageCacheMaxAge = time.Hour * 24 * 10 // 10 days
)

func (uc *UC) GetBinObjectData(
	ctx context.Context,
	auth *basedto.Auth,
	req *binobjectdto.GetBinObjectDataReq,
) (*binobjectdto.GetBinObjectDataResp, error) {
	binObject, err := uc.binObjectRepo.GetByID(ctx, uc.db, req.Type, req.ID,
		bunex.SelectWhere("bin_object.status = ?", base.BinObjectStatusActive),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	extraHeaders := map[string]string{}

	switch binObject.Type {
	case base.BinObjectTypeUserPhoto, base.BinObjectTypeProjectPhoto:
		extraHeaders["Cache-Control"] = fmt.Sprintf("public, max-age=%v",
			int(defaultImageCacheMaxAge.Seconds()))
	}

	return &binobjectdto.GetBinObjectDataResp{
		Data: &binobjectdto.BinObjectDataResp{
			ContentLength: int64(len(binObject.Data)),
			ContentType:   binObject.ContentType,
			Content:       io.NopCloser(bytes.NewReader(binObject.Data)),
			ExtraHeaders:  extraHeaders,
		},
	}, nil
}

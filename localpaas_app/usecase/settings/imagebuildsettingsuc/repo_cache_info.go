package imagebuildsettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imagebuildsettingsuc/imagebuildsettingsdto"
)

func (uc *UC) GetRepoCacheInfo(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagebuildsettingsdto.GetRepoCacheInfoReq,
) (*imagebuildsettingsdto.GetRepoCacheInfoResp, error) {
	// Supports scope global and project only
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectWhere("file.type = ?", base.FileTypeRepoCache),
		bunex.SelectWhere("file.storage_type = ?", base.FileStorageLocal),
		bunex.SelectWhere("file.deleted IS NOT TRUE"),
	}
	if !req.Scope.IsGlobalScope() {
		listOpts = append(listOpts,
			bunex.SelectWhere("file.scope = ?", req.Scope.ScopeType()),
			bunex.SelectWhere("file.object_id = ?", req.Scope.MainObjectID()),
		)
	}

	files, _, err := uc.FileRepo.List(ctx, uc.DB, nil, listOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &imagebuildsettingsdto.GetRepoCacheInfoResp{
		Data: imagebuildsettingsdto.TransformRepoCacheInfo(files),
	}, nil
}

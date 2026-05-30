package fileuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/fileuc/filedto"
)

func (uc *UC) ListFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *filedto.ListFileReq,
) (*filedto.ListFileResp, error) {
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectRelation("Storage"),
	}
	if len(req.Types) > 0 {
		listOpts = append(listOpts, bunex.SelectWhereIn("file.type IN (?)", req.Types...))
	}
	if len(req.Statuses) > 0 {
		listOpts = append(listOpts, bunex.SelectWhereIn("file.status IN (?)", req.Statuses...))
	}
	if len(req.Keys) > 0 {
		listOpts = append(listOpts, bunex.SelectWhereIn("file.key IN (?)", req.Keys...))
	}
	if len(req.StorageTypes) > 0 {
		listOpts = append(listOpts, bunex.SelectWhereIn("file.storage_type IN (?)", req.StorageTypes...))
	}
	if req.Search != "" {
		keyword := bunex.MakeLikeOpStr(req.Search, true)
		listOpts = append(listOpts,
			bunex.SelectWhereGroup(
				bunex.SelectWhere("file.name ILIKE ?", keyword),
			),
		)
	}
	if len(auth.AllowObjectIDs) > 0 {
		listOpts = append(listOpts,
			bunex.SelectWhereIn("file.id IN (?)", auth.AllowObjectIDs),
		)
	}

	files, paging, err := uc.fileRepo.List(ctx, uc.db, &req.Paging, listOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData, err := filedto.TransformFiles(files)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &filedto.ListFileResp{
		Meta: &basedto.ListMeta{Page: paging},
		Data: respData,
	}, nil
}

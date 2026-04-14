package repository

import (
	"context"

	"github.com/uptrace/bun"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

type AppTagRepo interface {
	List(ctx context.Context, db database.IDB, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.AppTag, *basedto.PagingMeta, error)

	UpsertMulti(ctx context.Context, db database.IDB, appTags []*entity.AppTag,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error

	DeleteAllByApps(ctx context.Context, db database.IDB, appIDs []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type appTagRepo struct {
}

func NewAppTagRepo() AppTagRepo {
	return &appTagRepo{}
}

func (repo *appTagRepo) List(ctx context.Context, db database.IDB, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.AppTag, *basedto.PagingMeta, error) {
	var appTags []*entity.AppTag
	query := db.NewSelect().Model(&appTags)
	query = bunex.ApplySelect(query, opts...)

	var pagingMeta *basedto.PagingMeta
	if paging != nil {
		pagingMeta = newPagingMeta(paging)

		// Counts the total first
		total, err := query.Count(ctx)
		if err != nil {
			return nil, nil, apperrors.Wrap(err)
		}
		pagingMeta.Total = total

		// Applies pagination
		query = bunex.ApplyPagination(query, paging)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, nil, wrapPaginationError(err, paging)
	}

	return appTags, pagingMeta, nil
}

func (repo *appTagRepo) UpsertMulti(ctx context.Context, db database.IDB, appTags []*entity.AppTag,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(appTags) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&appTags)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *appTagRepo) DeleteAllByApps(ctx context.Context, db database.IDB,
	appIDs []string, opts ...bunex.DeleteQueryOption) error {
	if len(appIDs) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.AppTag)(nil)).
		Where("app_id IN (?)", bun.List(appIDs))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *appTagRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	query := db.NewDelete().Model((*entity.AppTag)(nil)).ForceDelete()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

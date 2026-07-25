package repository

import (
	"context"

	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type TagRepo interface {
	List(ctx context.Context, db database.IDB, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.Tag, *basedto.PagingMeta, error)

	UpsertMulti(ctx context.Context, db database.IDB, tags []*entity.Tag,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error

	DeleteAllByObjects(ctx context.Context, db database.IDB, objectIDs []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type tagRepo struct {
}

func NewTagRepo() TagRepo {
	return &tagRepo{}
}

func (repo *tagRepo) List(ctx context.Context, db database.IDB, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.Tag, *basedto.PagingMeta, error) {
	var tags []*entity.Tag
	query := db.NewSelect().Model(&tags)
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

	return tags, pagingMeta, nil
}

func (repo *tagRepo) UpsertMulti(ctx context.Context, db database.IDB, tags []*entity.Tag,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(tags) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&tags)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *tagRepo) DeleteAllByObjects(ctx context.Context, db database.IDB,
	objectIDs []string, opts ...bunex.DeleteQueryOption) error {
	if len(objectIDs) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.Tag)(nil)).
		Where("object_id IN (?)", bun.List(objectIDs))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *tagRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return apperrors.NewArgumentInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.Tag)(nil)).ForceDelete().WhereAllWithDeleted()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

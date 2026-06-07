package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

type ResLinkRepo interface {
	Get(ctx context.Context, db database.IDB, srcType base.ResourceType, srcID string,
		dstType base.ResourceType, dstID string, opts ...bunex.SelectQueryOption) (*entity.ResLink, error)
	List(ctx context.Context, db database.IDB, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.ResLink, *basedto.PagingMeta, error)

	Insert(ctx context.Context, db database.IDB, resLink *entity.ResLink,
		opts ...bunex.InsertQueryOption) error
	InsertMulti(ctx context.Context, db database.IDB, resLinks []*entity.ResLink,
		opts ...bunex.InsertQueryOption) error
	Upsert(ctx context.Context, db database.IDB, resLink *entity.ResLink, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error
	UpsertMulti(ctx context.Context, db database.IDB, resLinks []*entity.ResLink, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error

	DeleteAllBySourceIDs(ctx context.Context, db database.IDB, sourceType base.SubjectType, sourceIDs []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type resLinkRepo struct {
}

func NewResLinkRepo() ResLinkRepo {
	return &resLinkRepo{}
}

func (repo *resLinkRepo) Get(ctx context.Context, db database.IDB, srcType base.ResourceType, srcID string,
	dstType base.ResourceType, dstID string, opts ...bunex.SelectQueryOption) (*entity.ResLink, error) {
	resLink := &entity.ResLink{}
	query := db.NewSelect().Model(resLink).
		Where("res_link.src_id = ?", srcID).
		Where("res_link.dst_id = ?", dstID)
	if srcType != "" {
		query = query.Where("res_link.src_type = ?", srcType)
	}
	if dstType != "" {
		query = query.Where("res_link.dst_type = ?", dstType)
	}
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if resLink == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("ResLink").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resLink, nil
}

func (repo *resLinkRepo) List(ctx context.Context, db database.IDB, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.ResLink, *basedto.PagingMeta, error) {
	var resLinks []*entity.ResLink
	query := db.NewSelect().Model(&resLinks)
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
	return resLinks, pagingMeta, nil
}

func (repo *resLinkRepo) Insert(ctx context.Context, db database.IDB, resLink *entity.ResLink,
	opts ...bunex.InsertQueryOption) error {
	return repo.InsertMulti(ctx, db, []*entity.ResLink{resLink}, opts...)
}

func (repo *resLinkRepo) InsertMulti(ctx context.Context, db database.IDB, resLinks []*entity.ResLink,
	opts ...bunex.InsertQueryOption) error {
	if len(resLinks) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&resLinks)
	query = bunex.ApplyInsert(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *resLinkRepo) Upsert(ctx context.Context, db database.IDB, resLink *entity.ResLink,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	return repo.UpsertMulti(ctx, db, []*entity.ResLink{resLink}, conflictCols, updateCols, opts...)
}

func (repo *resLinkRepo) UpsertMulti(ctx context.Context, db database.IDB, resLinks []*entity.ResLink,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(resLinks) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&resLinks)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *resLinkRepo) DeleteAllBySourceIDs(ctx context.Context, db database.IDB,
	sourceType base.SubjectType, sourceIDs []string, opts ...bunex.DeleteQueryOption) error {
	if len(sourceIDs) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.ResLink)(nil)).
		Where("res_link.src_type = ?", sourceType).
		Where("res_link.src_id IN (?)", bun.List(sourceIDs))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *resLinkRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return apperrors.NewParamInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.ResLink)(nil)).ForceDelete().WhereAllWithDeleted()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

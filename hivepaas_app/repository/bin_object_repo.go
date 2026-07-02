package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type BinObjectRepo interface {
	GetByID(ctx context.Context, db database.IDB, typ base.BinObjectType, id string,
		opts ...bunex.SelectQueryOption) (*entity.BinObject, error)
	List(ctx context.Context, db database.IDB, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.BinObject, *basedto.PagingMeta, error)
	ListByIDs(ctx context.Context, db database.IDB, ids []string,
		opts ...bunex.SelectQueryOption) ([]*entity.BinObject, error)

	Insert(ctx context.Context, db database.IDB, binObject *entity.BinObject,
		opts ...bunex.InsertQueryOption) error
	InsertMulti(ctx context.Context, db database.IDB, binaries []*entity.BinObject,
		opts ...bunex.InsertQueryOption) error
	Upsert(ctx context.Context, db database.IDB, binObject *entity.BinObject, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error
	UpsertMulti(ctx context.Context, db database.IDB, binaries []*entity.BinObject, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error
	Update(ctx context.Context, db database.IDB, binObject *entity.BinObject,
		opts ...bunex.UpdateQueryOption) error

	DeleteByIDs(ctx context.Context, db database.IDB, ids []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteHard(ctx context.Context, db database.IDB,
		opts ...bunex.DeleteQueryOption) error
}

type binObjectRepo struct {
}

func NewBinObjectRepo() BinObjectRepo {
	return &binObjectRepo{}
}

func (repo *binObjectRepo) GetByID(ctx context.Context, db database.IDB, typ base.BinObjectType, id string,
	opts ...bunex.SelectQueryOption) (*entity.BinObject, error) {
	binObject := &entity.BinObject{}
	query := db.NewSelect().Model(binObject).Where("bin_object.id = ?", id)
	if typ != "" {
		query = query.Where("bin_object.type = ?", typ)
	}
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if binObject == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("BinObject").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.New(err)
	}
	return binObject, nil
}

func (repo *binObjectRepo) List(ctx context.Context, db database.IDB, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.BinObject, *basedto.PagingMeta, error) {
	var binObjects []*entity.BinObject
	query := db.NewSelect().Model(&binObjects)
	query = bunex.ApplySelect(query, opts...)

	var pagingMeta *basedto.PagingMeta
	if paging != nil {
		pagingMeta = newPagingMeta(paging)

		// Counts the total first
		total, err := query.Count(ctx)
		if err != nil {
			return nil, nil, apperrors.New(err)
		}
		pagingMeta.Total = total

		// Applies pagination
		query = bunex.ApplyPagination(query, paging)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, nil, wrapPaginationError(err, paging)
	}
	return binObjects, pagingMeta, nil
}

func (repo *binObjectRepo) ListByIDs(ctx context.Context, db database.IDB, ids []string,
	opts ...bunex.SelectQueryOption) ([]*entity.BinObject, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var binObjects []*entity.BinObject
	query := db.NewSelect().Model(&binObjects).Where("bin_object.id IN (?)", bun.List(ids))
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return binObjects, nil
}

func (repo *binObjectRepo) Insert(ctx context.Context, db database.IDB, binObject *entity.BinObject,
	opts ...bunex.InsertQueryOption) error {
	return repo.InsertMulti(ctx, db, []*entity.BinObject{binObject}, opts...)
}

func (repo *binObjectRepo) InsertMulti(ctx context.Context, db database.IDB, binaries []*entity.BinObject,
	opts ...bunex.InsertQueryOption) error {
	if len(binaries) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&binaries)
	query = bunex.ApplyInsert(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *binObjectRepo) Upsert(ctx context.Context, db database.IDB, binObject *entity.BinObject,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	return repo.UpsertMulti(ctx, db, []*entity.BinObject{binObject}, conflictCols, updateCols, opts...)
}

func (repo *binObjectRepo) UpsertMulti(ctx context.Context, db database.IDB, binaries []*entity.BinObject,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(binaries) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&binaries)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *binObjectRepo) Update(ctx context.Context, db database.IDB, binObject *entity.BinObject,
	opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(binObject).WherePK()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

// NOTE: this UpdateMulti may not work properly with JSON columns defined as string in the entity struct
func (repo *binObjectRepo) UpdateMulti(ctx context.Context, db database.IDB, binaries []*entity.BinObject,
	opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(&binaries).Bulk()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *binObjectRepo) DeleteByIDs(ctx context.Context, db database.IDB, ids []string,
	opts ...bunex.DeleteQueryOption) error {
	if len(ids) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.BinObject)(nil)).
		Where("bin_object.id IN (?)", bun.List(ids))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *binObjectRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return apperrors.NewArgumentInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.BinObject)(nil)).ForceDelete().WhereAllWithDeleted()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

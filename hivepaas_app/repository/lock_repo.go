package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type LockRepo interface {
	GetByID(ctx context.Context, db database.IDB, id string,
		opts ...bunex.SelectQueryOption) (*entity.Lock, error)
	List(ctx context.Context, db database.IDB, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.Lock, *basedto.PagingMeta, error)

	Insert(ctx context.Context, db database.IDB, lock *entity.Lock,
		opts ...bunex.InsertQueryOption) error
	Upsert(ctx context.Context, db database.IDB, lock *entity.Lock, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error
	UpsertMulti(ctx context.Context, db database.IDB, locks []*entity.Lock, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error

	DeleteByIDs(ctx context.Context, db database.IDB, ids []string,
		opts ...bunex.DeleteQueryOption) error
}

type lockRepo struct {
}

func NewLockRepo() LockRepo {
	return &lockRepo{}
}

func (repo *lockRepo) GetByID(ctx context.Context, db database.IDB, id string,
	opts ...bunex.SelectQueryOption) (*entity.Lock, error) {
	lock := &entity.Lock{}
	query := db.NewSelect().Model(lock).Where("lock.id = ?", id)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if lock == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("Lock").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return lock, nil
}

func (repo *lockRepo) List(ctx context.Context, db database.IDB, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.Lock, *basedto.PagingMeta, error) {
	var locks []*entity.Lock
	query := db.NewSelect().Model(&locks)
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

	return locks, pagingMeta, nil
}

func (repo *lockRepo) Insert(ctx context.Context, db database.IDB, lock *entity.Lock,
	opts ...bunex.InsertQueryOption) error {
	query := db.NewInsert().Model(lock)
	query = bunex.ApplyInsert(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *lockRepo) Upsert(ctx context.Context, db database.IDB, lock *entity.Lock,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	return repo.UpsertMulti(ctx, db, []*entity.Lock{lock}, conflictCols, updateCols, opts...)
}

func (repo *lockRepo) UpsertMulti(ctx context.Context, db database.IDB, locks []*entity.Lock,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(locks) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&locks)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *lockRepo) DeleteByIDs(ctx context.Context, db database.IDB, ids []string,
	opts ...bunex.DeleteQueryOption) error {
	if len(ids) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.Lock)(nil)).
		Where("lock.id IN (?)", bun.List(ids))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

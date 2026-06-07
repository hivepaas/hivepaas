package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

type SysErrorRepo interface {
	GetByID(ctx context.Context, db database.IDB, id string,
		opts ...bunex.SelectQueryOption) (*entity.SysError, error)
	List(ctx context.Context, db database.IDB, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.SysError, *basedto.PagingMeta, error)

	Insert(ctx context.Context, db database.IDB, sysError *entity.SysError,
		opts ...bunex.InsertQueryOption) error
	InsertMulti(ctx context.Context, db database.IDB, sysErrors []*entity.SysError,
		opts ...bunex.InsertQueryOption) error

	Delete(ctx context.Context, db database.IDB, sysError *entity.SysError,
		opts ...bunex.DeleteQueryOption) error
	DeleteMulti(ctx context.Context, db database.IDB, sysErrors []*entity.SysError,
		opts ...bunex.DeleteQueryOption) error
	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type sysErrorRepo struct {
}

func NewSysErrorRepo() SysErrorRepo {
	return &sysErrorRepo{}
}

func (repo *sysErrorRepo) GetByID(ctx context.Context, db database.IDB, id string,
	opts ...bunex.SelectQueryOption) (*entity.SysError, error) {
	sysError := &entity.SysError{}
	query := db.NewSelect().Model(sysError).Where("sys_error.id = ?", id)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if sysError == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("SysError").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return sysError, nil
}

func (repo *sysErrorRepo) List(ctx context.Context, db database.IDB, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.SysError, *basedto.PagingMeta, error) {
	var sysErrors []*entity.SysError
	query := db.NewSelect().Model(&sysErrors)
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
	return sysErrors, pagingMeta, nil
}

func (repo *sysErrorRepo) Insert(ctx context.Context, db database.IDB, sysError *entity.SysError,
	opts ...bunex.InsertQueryOption) error {
	return repo.InsertMulti(ctx, db, []*entity.SysError{sysError}, opts...)
}

func (repo *sysErrorRepo) InsertMulti(ctx context.Context, db database.IDB, sysErrors []*entity.SysError,
	opts ...bunex.InsertQueryOption) error {
	if len(sysErrors) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&sysErrors)
	query = bunex.ApplyInsert(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *sysErrorRepo) Delete(ctx context.Context, db database.IDB, sysError *entity.SysError,
	opts ...bunex.DeleteQueryOption) error {
	return repo.DeleteMulti(ctx, db, []*entity.SysError{sysError}, opts...)
}

func (repo *sysErrorRepo) DeleteMulti(ctx context.Context, db database.IDB, sysErrors []*entity.SysError,
	opts ...bunex.DeleteQueryOption) error {
	if len(sysErrors) == 0 {
		return nil
	}
	query := db.NewDelete().Model(&sysErrors).WherePK()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *sysErrorRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return apperrors.NewParamInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.SysError)(nil)).ForceDelete() /* No soft delete */
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

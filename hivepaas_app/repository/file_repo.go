package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type FileRepo interface {
	GetByID(ctx context.Context, db database.IDB, id string,
		opts ...bunex.SelectQueryOption) (*entity.File, error)
	GetByName(ctx context.Context, db database.IDB, name string,
		opts ...bunex.SelectQueryOption) (*entity.File, error)
	GetByKey(ctx context.Context, db database.IDB, key string,
		opts ...bunex.SelectQueryOption) (*entity.File, error)

	List(ctx context.Context, db database.IDB, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.File, *basedto.PagingMeta, error)
	ListByIDs(ctx context.Context, db database.IDB, ids []string,
		opts ...bunex.SelectQueryOption) ([]*entity.File, error)

	Insert(ctx context.Context, db database.IDB, file *entity.File,
		opts ...bunex.InsertQueryOption) error
	InsertMulti(ctx context.Context, db database.IDB, files []*entity.File,
		opts ...bunex.InsertQueryOption) error
	Upsert(ctx context.Context, db database.IDB, file *entity.File,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	UpsertMulti(ctx context.Context, db database.IDB, files []*entity.File,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	Update(ctx context.Context, db database.IDB, file *entity.File,
		opts ...bunex.UpdateQueryOption) error

	DeleteAllByObjects(ctx context.Context, db database.IDB, scope base.ObjectScopeType, objectIDs []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type fileRepo struct {
}

func NewFileRepo() FileRepo {
	return &fileRepo{}
}

func (repo *fileRepo) GetByID(ctx context.Context, db database.IDB, id string,
	opts ...bunex.SelectQueryOption) (*entity.File, error) {
	file := &entity.File{}
	query := db.NewSelect().Model(file).Where("file.id = ?", id)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if file == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("File").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.New(err)
	}
	return file, nil
}

func (repo *fileRepo) GetByName(ctx context.Context, db database.IDB, name string,
	opts ...bunex.SelectQueryOption) (*entity.File, error) {
	file := &entity.File{}
	query := db.NewSelect().Model(file).Where("LOWER(file.name) = ?", strings.ToLower(name))
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if file == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("File").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.New(err)
	}
	return file, nil
}

func (repo *fileRepo) GetByKey(ctx context.Context, db database.IDB, key string,
	opts ...bunex.SelectQueryOption) (*entity.File, error) {
	file := &entity.File{}
	query := db.NewSelect().Model(file).Where("LOWER(file.key) = ?", key).Limit(1)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if file == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("File").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.New(err)
	}
	return file, nil
}

func (repo *fileRepo) List(ctx context.Context, db database.IDB, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.File, *basedto.PagingMeta, error) {
	var files []*entity.File
	query := db.NewSelect().Model(&files)
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

		// Apply pagination
		query = bunex.ApplyPagination(query, paging)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, nil, wrapPaginationError(err, paging)
	}

	return files, pagingMeta, nil
}

func (repo *fileRepo) ListByIDs(ctx context.Context, db database.IDB, ids []string,
	opts ...bunex.SelectQueryOption) ([]*entity.File, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var files []*entity.File
	query := db.NewSelect().Model(&files).Where("file.id IN (?)", bun.List(ids))
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return files, nil
}

func (repo *fileRepo) Insert(ctx context.Context, db database.IDB, file *entity.File,
	opts ...bunex.InsertQueryOption) error {
	return repo.InsertMulti(ctx, db, []*entity.File{file}, opts...)
}

func (repo *fileRepo) InsertMulti(ctx context.Context, db database.IDB, files []*entity.File,
	opts ...bunex.InsertQueryOption) error {
	if len(files) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&files)
	query = bunex.ApplyInsert(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *fileRepo) Upsert(ctx context.Context, db database.IDB, file *entity.File,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	return repo.UpsertMulti(ctx, db, []*entity.File{file}, conflictCols, updateCols, opts...)
}

func (repo *fileRepo) UpsertMulti(ctx context.Context, db database.IDB, files []*entity.File,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(files) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&files)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *fileRepo) Update(ctx context.Context, db database.IDB, file *entity.File,
	opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(file).WherePK()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *fileRepo) DeleteAllByObjects(ctx context.Context, db database.IDB,
	scope base.ObjectScopeType, objectIDs []string, opts ...bunex.DeleteQueryOption) error {
	if len(objectIDs) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.File)(nil)).
		Where("file.scope = ?", scope).
		Where("file.object_id IN (?)", bun.List(objectIDs))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *fileRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return apperrors.NewArgumentInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.File)(nil)).ForceDelete().WhereAllWithDeleted()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

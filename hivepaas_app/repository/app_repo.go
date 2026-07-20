package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type AppRepo interface {
	GetByID(ctx context.Context, db database.IDB, projectID, id string,
		opts ...bunex.SelectQueryOption) (*entity.App, error)
	GetByName(ctx context.Context, db database.IDB, projectID, name string,
		opts ...bunex.SelectQueryOption) (*entity.App, error)
	GetByGlobalKey(ctx context.Context, db database.IDB, projectID, gKey string,
		opts ...bunex.SelectQueryOption) (*entity.App, error)

	List(ctx context.Context, db database.IDB, projectID string, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.App, *basedto.PagingMeta, error)
	ListByIDs(ctx context.Context, db database.IDB, projectID string, ids []string,
		opts ...bunex.SelectQueryOption) ([]*entity.App, error)

	Upsert(ctx context.Context, db database.IDB, app *entity.App,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	UpsertMulti(ctx context.Context, db database.IDB, apps []*entity.App,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	Update(ctx context.Context, db database.IDB, app *entity.App,
		opts ...bunex.UpdateQueryOption) error

	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type appRepo struct {
}

func NewAppRepo() AppRepo {
	return &appRepo{}
}

func (repo *appRepo) GetByID(ctx context.Context, db database.IDB, projectID, id string,
	opts ...bunex.SelectQueryOption) (*entity.App, error) {
	app := &entity.App{}
	query := db.NewSelect().Model(app).Where("app.id = ?", id)
	if projectID != "" {
		query = query.Where("app.project_id = ?", projectID)
	}
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if app == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("App").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return app, nil
}

func (repo *appRepo) GetByName(ctx context.Context, db database.IDB, projectID, name string,
	opts ...bunex.SelectQueryOption) (*entity.App, error) {
	app := &entity.App{}
	query := db.NewSelect().Model(app).Where("LOWER(app.name) = ?", strings.ToLower(name))
	if projectID != "" {
		query = query.Where("app.project_id = ?", projectID)
	}
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if app == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("App").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return app, nil
}

func (repo *appRepo) GetByGlobalKey(ctx context.Context, db database.IDB, projectID, gKey string,
	opts ...bunex.SelectQueryOption) (*entity.App, error) {
	app := &entity.App{}
	query := db.NewSelect().Model(app).Where("app.global_key = ?", gKey).Limit(1)
	if projectID != "" {
		query = query.Where("app.project_id = ?", projectID)
	}
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if app == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("App").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return app, nil
}

func (repo *appRepo) List(ctx context.Context, db database.IDB, projectID string, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.App, *basedto.PagingMeta, error) {
	var apps []*entity.App
	query := db.NewSelect().Model(&apps)
	if projectID != "" {
		query = query.Where("app.project_id = ?", projectID)
	}
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

	return apps, pagingMeta, nil
}

func (repo *appRepo) ListByIDs(ctx context.Context, db database.IDB, projectID string, ids []string,
	opts ...bunex.SelectQueryOption) ([]*entity.App, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var apps []*entity.App
	query := db.NewSelect().Model(&apps).Where("app.id IN (?)", bun.List(ids))
	if projectID != "" {
		query = query.Where("app.project_id = ?", projectID)
	}
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return apps, nil
}

func (repo *appRepo) Upsert(ctx context.Context, db database.IDB, app *entity.App,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	return repo.UpsertMulti(ctx, db, []*entity.App{app}, conflictCols, updateCols, opts...)
}

func (repo *appRepo) UpsertMulti(ctx context.Context, db database.IDB, apps []*entity.App,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(apps) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&apps)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *appRepo) Update(ctx context.Context, db database.IDB, app *entity.App,
	opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(app).WherePK()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *appRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return apperrors.NewArgumentInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.App)(nil)).ForceDelete().WhereAllWithDeleted()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

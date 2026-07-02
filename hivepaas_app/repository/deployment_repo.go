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

type DeploymentRepo interface {
	GetByID(ctx context.Context, db database.IDB, appID, id string,
		opts ...bunex.SelectQueryOption) (*entity.Deployment, error)
	List(ctx context.Context, db database.IDB, appID string, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.Deployment, *basedto.PagingMeta, error)
	ListByIDs(ctx context.Context, db database.IDB, appID string, ids []string,
		opts ...bunex.SelectQueryOption) ([]*entity.Deployment, error)

	Upsert(ctx context.Context, db database.IDB, deployment *entity.Deployment, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error
	UpsertMulti(ctx context.Context, db database.IDB, deployments []*entity.Deployment, conflictCols, updateCols []string,
		opts ...bunex.InsertQueryOption) error
	Update(ctx context.Context, db database.IDB, deployment *entity.Deployment,
		opts ...bunex.UpdateQueryOption) error

	DeleteAllByApps(ctx context.Context, db database.IDB, appIDs []string, opts ...bunex.DeleteQueryOption) error
	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type deploymentRepo struct {
}

func NewDeploymentRepo() DeploymentRepo {
	return &deploymentRepo{}
}

func (repo *deploymentRepo) GetByID(ctx context.Context, db database.IDB, appID, id string,
	opts ...bunex.SelectQueryOption) (*entity.Deployment, error) {
	deployment := &entity.Deployment{}
	query := db.NewSelect().Model(deployment).Where("deployment.id = ?", id)
	if appID != "" {
		query = query.Where("deployment.app_id = ?", appID)
	}
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if deployment == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("Deployment").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.New(err)
	}
	return deployment, nil
}

func (repo *deploymentRepo) List(ctx context.Context, db database.IDB, appID string, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.Deployment, *basedto.PagingMeta, error) {
	var deployments []*entity.Deployment
	query := db.NewSelect().Model(&deployments)
	if appID != "" {
		query = query.Where("deployment.app_id = ?", appID)
	}
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
	return deployments, pagingMeta, nil
}

func (repo *deploymentRepo) ListByIDs(ctx context.Context, db database.IDB, appID string, ids []string,
	opts ...bunex.SelectQueryOption) ([]*entity.Deployment, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var deployments []*entity.Deployment
	query := db.NewSelect().Model(&deployments).Where("deployment.id IN (?)", bun.List(ids))
	if appID != "" {
		query = query.Where("deployment.app_id = ?", appID)
	}
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return deployments, nil
}

func (repo *deploymentRepo) Upsert(ctx context.Context, db database.IDB, deployment *entity.Deployment,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	return repo.UpsertMulti(ctx, db, []*entity.Deployment{deployment}, conflictCols, updateCols, opts...)
}

func (repo *deploymentRepo) UpsertMulti(ctx context.Context, db database.IDB, deployments []*entity.Deployment,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(deployments) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&deployments)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *deploymentRepo) Update(ctx context.Context, db database.IDB, deployment *entity.Deployment,
	opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(deployment).WherePK()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *deploymentRepo) DeleteAllByApps(ctx context.Context, db database.IDB, appIDs []string,
	opts ...bunex.DeleteQueryOption) error {
	if len(appIDs) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.Deployment)(nil)).
		Where("app_id IN (?)", bun.List(appIDs))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (repo *deploymentRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return apperrors.NewArgumentInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.Deployment)(nil)).ForceDelete().WhereAllWithDeleted()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

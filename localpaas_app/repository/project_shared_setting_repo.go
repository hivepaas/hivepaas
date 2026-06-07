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

type ProjectSharedSettingRepo interface {
	Get(ctx context.Context, db database.IDB, projectID, settingID string,
		opts ...bunex.SelectQueryOption) (*entity.ProjectSharedSetting, error)
	List(ctx context.Context, db database.IDB, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.ProjectSharedSetting, *basedto.PagingMeta, error)

	UpsertMulti(ctx context.Context, db database.IDB, projectSharedSettings []*entity.ProjectSharedSetting,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	Update(ctx context.Context, db database.IDB, projectSharedSetting *entity.ProjectSharedSetting,
		opts ...bunex.UpdateQueryOption) error

	DeleteAllBySetting(ctx context.Context, db database.IDB, settingID string,
		opts ...bunex.DeleteQueryOption) error
	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type projectSharedSettingRepo struct {
}

func NewProjectSharedSettingRepo() ProjectSharedSettingRepo {
	return &projectSharedSettingRepo{}
}

func (repo *projectSharedSettingRepo) Get(ctx context.Context, db database.IDB, projectID, settingID string,
	opts ...bunex.SelectQueryOption) (*entity.ProjectSharedSetting, error) {
	projectSharedSetting := &entity.ProjectSharedSetting{}
	query := db.NewSelect().Model(projectSharedSetting).
		Where("project_shared_setting.project_id = ?", projectID).
		Where("project_shared_setting.setting_id = ?", settingID)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if projectSharedSetting == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("ProjectSharedSetting").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return projectSharedSetting, nil
}

func (repo *projectSharedSettingRepo) List(ctx context.Context, db database.IDB, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.ProjectSharedSetting, *basedto.PagingMeta, error) {
	var projectSharedSettings []*entity.ProjectSharedSetting
	query := db.NewSelect().Model(&projectSharedSettings)
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

	return projectSharedSettings, pagingMeta, nil
}

func (repo *projectSharedSettingRepo) UpsertMulti(ctx context.Context, db database.IDB,
	projectSharedSettings []*entity.ProjectSharedSetting, conflictCols, updateCols []string,
	opts ...bunex.InsertQueryOption) error {
	if len(projectSharedSettings) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&projectSharedSettings)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *projectSharedSettingRepo) Update(ctx context.Context, db database.IDB,
	projectSharedSetting *entity.ProjectSharedSetting, opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(projectSharedSetting).WherePK()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *projectSharedSettingRepo) DeleteAllBySetting(ctx context.Context, db database.IDB,
	settingID string, opts ...bunex.DeleteQueryOption) error {
	query := db.NewDelete().Model((*entity.ProjectSharedSetting)(nil)).
		Where("project_shared_setting.setting_id = ?", settingID)
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *projectSharedSettingRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return apperrors.NewParamInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.ProjectSharedSetting)(nil)).ForceDelete().WhereAllWithDeleted()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

type SharedSettingRepo interface {
	Get(ctx context.Context, db database.IDB, objectID, settingID string,
		opts ...bunex.SelectQueryOption) (*entity.SharedSetting, error)
	List(ctx context.Context, db database.IDB, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.SharedSetting, *basedto.PagingMeta, error)

	UpsertMulti(ctx context.Context, db database.IDB, sharedSettings []*entity.SharedSetting,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	Update(ctx context.Context, db database.IDB, sharedSetting *entity.SharedSetting,
		opts ...bunex.UpdateQueryOption) error

	DeleteAllBySetting(ctx context.Context, db database.IDB, settingID string,
		opts ...bunex.DeleteQueryOption) error
	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type sharedSettingRepo struct {
}

func NewSharedSettingRepo() SharedSettingRepo {
	return &sharedSettingRepo{}
}

func (repo *sharedSettingRepo) Get(ctx context.Context, db database.IDB, objectID, settingID string,
	opts ...bunex.SelectQueryOption) (*entity.SharedSetting, error) {
	sharedSetting := &entity.SharedSetting{}
	query := db.NewSelect().Model(sharedSetting).
		Where("shared_setting.object_id = ?", objectID).
		Where("shared_setting.setting_id = ?", settingID)
	query = bunex.ApplySelect(query, opts...)

	err := query.Scan(ctx)
	if sharedSetting == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, hperrors.NewNotFound("SharedSetting").WithCause(err)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return sharedSetting, nil
}

func (repo *sharedSettingRepo) List(ctx context.Context, db database.IDB, paging *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.SharedSetting, *basedto.PagingMeta, error) {
	var sharedSettings []*entity.SharedSetting
	query := db.NewSelect().Model(&sharedSettings)
	query = bunex.ApplySelect(query, opts...)

	var pagingMeta *basedto.PagingMeta
	if paging != nil {
		pagingMeta = newPagingMeta(paging)

		// Counts the total first
		total, err := query.Count(ctx)
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
		pagingMeta.Total = total

		// Applies pagination
		query = bunex.ApplyPagination(query, paging)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, nil, wrapPaginationError(err, paging)
	}

	return sharedSettings, pagingMeta, nil
}

func (repo *sharedSettingRepo) UpsertMulti(ctx context.Context, db database.IDB,
	sharedSettings []*entity.SharedSetting, conflictCols, updateCols []string,
	opts ...bunex.InsertQueryOption) error {
	if len(sharedSettings) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&sharedSettings)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *sharedSettingRepo) Update(ctx context.Context, db database.IDB,
	sharedSetting *entity.SharedSetting, opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(sharedSetting).WherePK()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *sharedSettingRepo) DeleteAllBySetting(ctx context.Context, db database.IDB,
	settingID string, opts ...bunex.DeleteQueryOption) error {
	query := db.NewDelete().Model((*entity.SharedSetting)(nil)).
		Where("shared_setting.setting_id = ?", settingID)
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (repo *sharedSettingRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return hperrors.NewArgumentInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.SharedSetting)(nil)).ForceDelete().WhereAllWithDeleted()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

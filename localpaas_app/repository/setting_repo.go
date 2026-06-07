package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/tiendc/gofn"
	"github.com/uptrace/bun"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

type SettingRepo interface {
	GetByID(ctx context.Context, db database.IDB, scope *base.ObjectScope, typ base.SettingType,
		id string, requireActive bool, opts ...bunex.SelectQueryOption) (*entity.Setting, error)
	GetByKind(ctx context.Context, db database.IDB, scope *base.ObjectScope, typ base.SettingType,
		kind string, requireActive bool, opts ...bunex.SelectQueryOption) (*entity.Setting, error)
	GetByName(ctx context.Context, db database.IDB, scope *base.ObjectScope, typ base.SettingType,
		name string, requireActive bool, opts ...bunex.SelectQueryOption) (*entity.Setting, error)
	GetSingle(ctx context.Context, db database.IDB, scope *base.ObjectScope, typ base.SettingType,
		requireActive bool, opts ...bunex.SelectQueryOption) (*entity.Setting, error)

	List(ctx context.Context, db database.IDB, scope *base.ObjectScope, paging *basedto.Paging,
		opts ...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error)
	ListByIDs(ctx context.Context, db database.IDB, scope *base.ObjectScope, ids []string, requireActive bool,
		opts ...bunex.SelectQueryOption) ([]*entity.Setting, error)

	// EnsureUnique makes sure there is at most one active setting in the given scope.
	// Inherited/imported settings are not taken into account.
	EnsureUnique(ctx context.Context, db database.IDB, scope *base.ObjectScope, typ base.SettingType,
		opts ...bunex.SelectQueryOption) error

	Insert(ctx context.Context, db database.IDB, setting *entity.Setting,
		opts ...bunex.InsertQueryOption) error
	InsertMulti(ctx context.Context, db database.IDB, settings []*entity.Setting,
		opts ...bunex.InsertQueryOption) error
	Upsert(ctx context.Context, db database.IDB, setting *entity.Setting,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	UpsertMulti(ctx context.Context, db database.IDB, settings []*entity.Setting,
		conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error
	Update(ctx context.Context, db database.IDB, setting *entity.Setting,
		opts ...bunex.UpdateQueryOption) error
	UpdateClearDefaultFlag(ctx context.Context, db database.IDB, scope *base.ObjectScope, typ base.SettingType,
		kind *string, exceptID string, opts ...bunex.UpdateQueryOption) error

	DeleteAllByObjects(ctx context.Context, db database.IDB, scope base.ObjectScopeType, objectIDs []string,
		opts ...bunex.DeleteQueryOption) error
	DeleteHard(ctx context.Context, db database.IDB, opts ...bunex.DeleteQueryOption) error
}

type settingRepo struct {
	appRepo AppRepo
}

func NewSettingRepo(appRepo AppRepo) SettingRepo {
	return &settingRepo{
		appRepo: appRepo,
	}
}

func (repo *settingRepo) GetByID(ctx context.Context, db database.IDB, scope *base.ObjectScope,
	typ base.SettingType, id string, requireActive bool,
	opts ...bunex.SelectQueryOption) (*entity.Setting, error) {
	opts = repo.applyFilter(opts, typ, id, requireActive)
	return repo.get(ctx, db, scope, opts...)
}

func (repo *settingRepo) GetByKind(ctx context.Context, db database.IDB, scope *base.ObjectScope,
	typ base.SettingType, kind string, requireActive bool,
	opts ...bunex.SelectQueryOption) (*entity.Setting, error) {
	if kind == "" {
		return nil, nil
	}
	opts = repo.applyFilter(opts, typ, "", requireActive)
	opts = repo.applyNameAndKindFilter(opts, "", kind)
	return repo.get(ctx, db, scope, opts...)
}

func (repo *settingRepo) GetByName(ctx context.Context, db database.IDB, scope *base.ObjectScope,
	typ base.SettingType, name string, requireActive bool,
	opts ...bunex.SelectQueryOption) (*entity.Setting, error) {
	if name == "" {
		return nil, nil
	}
	opts = repo.applyFilter(opts, typ, "", requireActive)
	opts = repo.applyNameAndKindFilter(opts, name, "")
	return repo.get(ctx, db, scope, opts...)
}

func (repo *settingRepo) get(ctx context.Context, db database.IDB, scope *base.ObjectScope,
	opts ...bunex.SelectQueryOption) (*entity.Setting, error) {
	theOpts := opts
	if scope != nil {
		// Query project ID for the app if it's not given
		if scope.AppID != "" && scope.ProjectID == "" {
			app, err := repo.appRepo.GetByID(ctx, db, "", scope.AppID,
				bunex.SelectColumns("project_id", "parent_id"))
			if err != nil {
				return nil, apperrors.Wrap(err)
			}
			scope.ParentAppID = app.ParentID
			scope.ProjectID = app.ProjectID
		}

		switch {
		case scope.AppID != "":
			theOpts = repo.applyAppFilter(theOpts, scope.AppID, scope.ParentAppID, scope.ProjectID)
		case scope.ProjectID != "":
			theOpts = repo.applyProjectFilter(theOpts, scope.ProjectID)
		case scope.UserID != "":
			theOpts = repo.applyUserFilter(theOpts, scope.UserID)
		default:
			theOpts = repo.applyGlobalFilter(theOpts)
		}
	}

	setting := &entity.Setting{}
	query := db.NewSelect().Model(setting)
	query = bunex.ApplySelect(query, theOpts...)

	err := query.Scan(ctx)
	if setting == nil || errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("Setting").WithCause(err)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if hasChange, _ := repo.updateExpiredSetting(ctx, db, setting); hasChange {
		return repo.get(ctx, db, scope, opts...)
	}
	return setting, nil
}

//nolint:gocognit
func (repo *settingRepo) GetSingle(ctx context.Context, db database.IDB, scope *base.ObjectScope,
	typ base.SettingType, requireActive bool, opts ...bunex.SelectQueryOption) (*entity.Setting, error) {
	opts = repo.applyFilter(opts, typ, "", requireActive)

	// Global scope is the special case
	if scope != nil && scope.IsGlobalScope() {
		opts = repo.applyGlobalFilter(opts)
		setting, err := repo.get(ctx, db, scope, opts...)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		return setting, nil
	}

	// For other scopes, we need to list satisfied settings upto the global scope,
	// then return the first matching one in the order of the scope upto global.
	settings, _, err := repo.List(ctx, db, scope, nil, opts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if len(settings) == 0 {
		return nil, apperrors.NewNotFound("Setting")
	}
	if len(settings) == 1 || scope == nil {
		return settings[0], nil
	}

	var parentSetting, projectSetting, globalSetting *entity.Setting
	for _, setting := range settings {
		switch {
		case scope.AppID != "":
			if setting.ObjectID == scope.AppID { // app's direct setting has the highest priority
				return setting, nil
			}
			if parentSetting == nil && setting.ObjectID != "" && setting.ObjectID == scope.ParentAppID {
				parentSetting = setting
				continue
			}
			if projectSetting == nil && setting.ObjectID != "" && setting.ObjectID == scope.ProjectID {
				projectSetting = setting
				continue
			}
		case scope.ProjectID != "":
			if setting.ObjectID == scope.ProjectID { // project's direct setting has the highest priority
				return setting, nil
			}
		case scope.UserID != "":
			if setting.ObjectID == scope.UserID { // // user's direct setting has the highest priority
				return setting, nil
			}
		default:
			return setting, nil
		}

		if globalSetting == nil && setting.ObjectID == "" {
			globalSetting = setting
			continue
		}
	}
	setting := gofn.Coalesce(parentSetting, projectSetting, globalSetting)
	if setting != nil {
		return setting, nil
	}
	return nil, apperrors.NewNotFound("Setting")
}

func (repo *settingRepo) List(ctx context.Context, db database.IDB, scope *base.ObjectScope,
	paging *basedto.Paging, opts ...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	theOpts := opts
	if scope != nil {
		// Query project ID for the app if it's not given
		if scope.AppID != "" && scope.ProjectID == "" {
			app, err := repo.appRepo.GetByID(ctx, db, "", scope.AppID,
				bunex.SelectColumns("project_id", "parent_id"))
			if err != nil {
				return nil, nil, apperrors.Wrap(err)
			}
			scope.ParentAppID = app.ParentID
			scope.ProjectID = app.ProjectID
		}

		switch {
		case scope.AppID != "":
			theOpts = repo.applyAppFilter(theOpts, scope.AppID, scope.ParentAppID, scope.ProjectID)
		case scope.ProjectID != "":
			theOpts = repo.applyProjectFilter(theOpts, scope.ProjectID)
		case scope.UserID != "":
			theOpts = repo.applyUserFilter(theOpts, scope.UserID)
		default:
			theOpts = repo.applyGlobalFilter(theOpts)
		}
	}

	var settings []*entity.Setting
	query := db.NewSelect().Model(&settings)
	query = bunex.ApplySelect(query, theOpts...)

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

	if hasChange, _ := repo.updateExpiredSettings(ctx, db, settings); hasChange {
		return repo.List(ctx, db, scope, paging, opts...)
	}
	return settings, pagingMeta, nil
}

func (repo *settingRepo) ListByIDs(ctx context.Context, db database.IDB, scope *base.ObjectScope,
	ids []string, requireActive bool, opts ...bunex.SelectQueryOption) ([]*entity.Setting, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	opts = append(opts, bunex.SelectWhere("setting.id IN (?)", bun.List(ids)))
	opts = repo.applyFilter(opts, "", "", requireActive)
	settings, _, err := repo.List(ctx, db, scope, nil, opts...)
	return settings, apperrors.Wrap(err)
}

func (repo *settingRepo) applyFilter(opts []bunex.SelectQueryOption, typ base.SettingType, id string,
	requireActive bool) []bunex.SelectQueryOption {
	if typ != "" {
		opts = append(opts, bunex.SelectWhere("setting.type = ?", typ))
	}
	if id != "" {
		opts = append(opts, bunex.SelectWhere("setting.id = ?", id))
	}
	if requireActive {
		opts = append(opts, bunex.SelectWhere("setting.status = ?", base.SettingStatusActive))
	}
	return opts
}

func (repo *settingRepo) applyNameAndKindFilter(opts []bunex.SelectQueryOption,
	name, kind string) []bunex.SelectQueryOption {
	if name != "" {
		opts = append(opts, bunex.SelectWhere("LOWER(setting.name) = ?", strings.ToLower(name)))
	}
	if kind != "" {
		opts = append(opts, bunex.SelectWhere("setting.kind = ?", kind))
	}
	return opts
}

func (repo *settingRepo) applyAppFilter(opts []bunex.SelectQueryOption,
	appID, parentAppID, projectID string) []bunex.SelectQueryOption {
	if projectID != "" {
		opts = append(opts,
			bunex.SelectJoin("LEFT JOIN project_shared_settings pss ON pss.setting_id = setting.id"),
			bunex.SelectWhereGroup(
				bunex.SelectWhereIf(appID != "", "setting.object_id = ?", appID),
				bunex.SelectWhereOrIf(parentAppID != "", "setting.object_id = ?", parentAppID),
				bunex.SelectWhereOr("setting.object_id = ?", projectID),
				bunex.SelectWhereOr("(setting.object_id IS NULL AND setting.avail_in_projects = TRUE)"),
				bunex.SelectWhereOr("(setting.object_id IS NULL AND pss.project_id = ? AND pss.deleted_at IS NULL)",
					projectID),
			),
		)
	} else if appID != "" {
		opts = append(opts,
			bunex.SelectWhereGroup(
				bunex.SelectWhere("setting.object_id = ?", appID),
				bunex.SelectWhereOrIf(parentAppID != "", "setting.object_id = ?", parentAppID),
			),
		)
	}
	return opts
}

func (repo *settingRepo) applyProjectFilter(opts []bunex.SelectQueryOption,
	projectID string) []bunex.SelectQueryOption {
	opts = append(opts,
		bunex.SelectJoin("LEFT JOIN project_shared_settings pss ON pss.setting_id = setting.id"),
		bunex.SelectWhereGroup(
			bunex.SelectWhere("setting.object_id = ?", projectID),
			bunex.SelectWhereOr("(setting.object_id IS NULL AND setting.avail_in_projects = TRUE)"),
			bunex.SelectWhereOr("(setting.object_id IS NULL AND pss.project_id = ? AND pss.deleted_at IS NULL)",
				projectID),
		),
	)
	return opts
}

func (repo *settingRepo) applyUserFilter(opts []bunex.SelectQueryOption, userID string) []bunex.SelectQueryOption {
	opts = append(opts, bunex.SelectWhere("setting.object_id = ?", userID))
	return opts
}

func (repo *settingRepo) applyGlobalFilter(opts []bunex.SelectQueryOption) []bunex.SelectQueryOption {
	opts = append(opts, bunex.SelectWhere("setting.object_id IS NULL"))
	return opts
}

func (repo *settingRepo) EnsureUnique(ctx context.Context, db database.IDB, scope *base.ObjectScope,
	typ base.SettingType, opts ...bunex.SelectQueryOption) error {
	query := db.NewSelect().Model((*entity.Setting)(nil)).
		Where("setting.type = ?", typ).
		Where("setting.status = ?", base.SettingStatusActive)
	if scope != nil {
		switch scope.ScopeType() {
		case base.ObjectScopeGlobal:
			query = query.Where("setting.object_id IS NULL")
		case base.ObjectScopeProject:
			query = query.Where("setting.object_id = ?", scope.ProjectID)
		case base.ObjectScopeApp:
			query = query.Where("setting.object_id = ?", scope.AppID)
		case base.ObjectScopeUser:
			query = query.Where("setting.object_id = ?", scope.UserID)
		default:
			// Do nothing
		}
	}
	query = bunex.ApplySelect(query, opts...)

	count, err := query.Count(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return apperrors.Wrap(err)
	}
	if count > 1 {
		return apperrors.NewAlreadyExist("Setting")
	}
	return nil
}

func (repo *settingRepo) Insert(ctx context.Context, db database.IDB, setting *entity.Setting,
	opts ...bunex.InsertQueryOption) error {
	return repo.InsertMulti(ctx, db, []*entity.Setting{setting}, opts...)
}

func (repo *settingRepo) InsertMulti(ctx context.Context, db database.IDB, settings []*entity.Setting,
	opts ...bunex.InsertQueryOption) error {
	if len(settings) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&settings)
	query = bunex.ApplyInsert(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *settingRepo) Upsert(ctx context.Context, db database.IDB, setting *entity.Setting,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	return repo.UpsertMulti(ctx, db, []*entity.Setting{setting}, conflictCols, updateCols, opts...)
}

func (repo *settingRepo) UpsertMulti(ctx context.Context, db database.IDB, settings []*entity.Setting,
	conflictCols, updateCols []string, opts ...bunex.InsertQueryOption) error {
	if len(settings) == 0 {
		return nil
	}
	query := db.NewInsert().Model(&settings)
	query = bunex.ApplyInsert(query, opts...)
	query = bunex.ApplyUpsert(query, conflictCols, updateCols)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *settingRepo) Update(ctx context.Context, db database.IDB, setting *entity.Setting,
	opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model(setting).WherePK()
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *settingRepo) UpdateClearDefaultFlag(ctx context.Context, db database.IDB, scope *base.ObjectScope,
	typ base.SettingType, kind *string, exceptID string, opts ...bunex.UpdateQueryOption) error {
	query := db.NewUpdate().Model((*entity.Setting)(nil)).
		Where("setting.type = ?", typ).
		Where("setting.is_default = true").
		Set("is_default = false")
	if kind != nil {
		if *kind == "" {
			query = query.Where("(setting.kind IS NULL OR setting.kind = '')")
		} else {
			query = query.Where("setting.kind = ?", *kind)
		}
	}
	if exceptID != "" {
		query = query.Where("setting.id != ?", exceptID)
	}
	switch {
	case scope.AppID != "":
		query = query.Where("setting.object_id = ?", scope.AppID)
	case scope.ProjectID != "":
		query = query.Where("setting.object_id = ?", scope.ProjectID)
	case scope.UserID != "":
		query = query.Where("setting.object_id = ?", scope.UserID)
	default:
		query = query.Where("setting.object_id IS NULL")
	}
	query = bunex.ApplyUpdate(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *settingRepo) updateExpiredSetting(ctx context.Context, db database.IDB,
	setting *entity.Setting) (bool, error) {
	if setting == nil {
		return false, nil
	}
	return repo.updateExpiredSettings(ctx, db, []*entity.Setting{setting})
}

func (repo *settingRepo) updateExpiredSettings(ctx context.Context, db database.IDB,
	settings []*entity.Setting) (hasChange bool, err error) {
	for _, setting := range settings {
		if setting.IsStatusDirty() {
			hasChange = true
			break
		}
	}
	if !hasChange {
		return false, nil
	}
	query := db.NewUpdate().Model((*entity.Setting)(nil)).
		Set("status = ?", base.SettingStatusExpired).
		Where("status = ? AND expire_at < NOW()", base.SettingStatusActive)

	_, err = query.Exec(ctx)
	if err != nil {
		return hasChange, apperrors.Wrap(err)
	}
	return hasChange, nil
}

func (repo *settingRepo) DeleteAllByObjects(ctx context.Context, db database.IDB,
	scope base.ObjectScopeType, objectIDs []string, opts ...bunex.DeleteQueryOption) error {
	if len(objectIDs) == 0 {
		return nil
	}
	query := db.NewDelete().Model((*entity.Setting)(nil)).
		Where("setting.scope = ?", scope).
		Where("setting.object_id IN (?)", bun.List(objectIDs))
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (repo *settingRepo) DeleteHard(ctx context.Context, db database.IDB,
	opts ...bunex.DeleteQueryOption) error {
	if len(opts) == 0 {
		return apperrors.NewParamInvalid("opts").WithMsgLog("DeleteHard requires at least one condition")
	}
	query := db.NewDelete().Model((*entity.Setting)(nil)).ForceDelete().WhereAllWithDeleted()
	query = bunex.ApplyDelete(query, opts...)

	_, err := query.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

package settings

import (
	"context"
	"errors"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/entityutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/strutil"
)

type BaseSettingReq struct {
	Type  base.SettingType  `json:"-" mapstructure:"-"`
	Kind  string            `json:"-" mapstructure:"-"`
	Scope *base.ObjectScope `json:"-" mapstructure:"-"`
}

type BaseSettingResp struct {
	ID              string             `json:"id"`
	Type            base.SettingType   `json:"type"`
	Name            string             `json:"name"`
	Kind            string             `json:"kind"`
	Status          base.SettingStatus `json:"status"`
	Size            int64              `json:"size,omitempty"`
	Inherited       bool               `json:"inherited,omitempty"`
	AvailInProjects bool               `json:"availableInProjects,omitempty"`
	Default         bool               `json:"default,omitempty"`
	UpdateVer       int                `json:"updateVer"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	ExpireAt  *time.Time `json:"expireAt,omitempty" copy:",nilonzero"`
}

type BaseSettingData struct {
	ScopeProject *entity.Project
	ScopeApp     *entity.App
	ScopeUser    *entity.User
}

func (uc *BaseUC) loadSettingScopeData(
	ctx context.Context,
	db database.IDB,
	req *BaseSettingReq,
	data *BaseSettingData,
) (err error) {
	requireActive := !req.Scope.NotRequireActive
	switch req.Scope.ScopeType() {
	case base.ObjectScopeGlobal:
		return nil

	case base.ObjectScopeProject:
		data.ScopeProject, err = uc.ProjectService.LoadProject(ctx, db, req.Scope.ProjectID, requireActive,
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		)
		if err != nil {
			return apperrors.New(err)
		}

	case base.ObjectScopeApp:
		data.ScopeApp, err = uc.AppService.LoadApp(ctx, db, req.Scope.ProjectID, req.Scope.AppID,
			requireActive, requireActive,
			bunex.SelectRelation("Project",
				bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
			),
			bunex.SelectRelation("ParentApp",
				bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			),
			bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		)
		if err != nil {
			return apperrors.New(err)
		}
		data.ScopeProject = data.ScopeApp.Project
		req.Scope.ProjectID = data.ScopeApp.ProjectID
		req.Scope.ParentAppID = data.ScopeApp.ParentID

	case base.ObjectScopeUser:
		data.ScopeUser, err = uc.UserService.LoadUserEx(ctx, db, req.Scope.UserID, requireActive)
		if err != nil {
			return apperrors.New(err)
		}
	}

	return nil
}

func (uc *BaseUC) loadSettingByID(
	ctx context.Context,
	db database.IDB,
	req *BaseSettingReq,
	id string,
	requireActive bool,
	opts ...bunex.SelectQueryOption,
) (setting *entity.Setting, err error) {
	if req.Kind != "" {
		opts = append(opts, bunex.SelectWhere("setting.kind = ?", req.Kind))
	}
	setting, err = uc.SettingRepo.GetByID(ctx, db, req.Scope, req.Type, id, requireActive, opts...)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return setting, nil
}

func (uc *BaseUC) checkNameConflict(
	ctx context.Context,
	db database.IDB,
	req *BaseSettingReq,
	name string,
) (err error) {
	if name == "" {
		return nil
	}
	setting, err := uc.SettingRepo.GetByName(ctx, db, req.Scope, req.Type, name, false)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.New(err)
	}
	if setting != nil {
		return apperrors.NewAlreadyExist(strutil.ToPascalCase(string(req.Type))).
			WithMsgLog("%s '%s' already exists", req.Type, setting.Name)
	}
	return nil
}

func (uc *BaseUC) checkRefObjectsExistence(
	ctx context.Context,
	db database.IDB,
	req *BaseSettingReq,
	refIDs *entity.RefObjectIDs,
	requireActive bool,
) (err error) {
	if refIDs == nil {
		return nil
	}
	err = uc.checkRefSettingsExistence(ctx, db, req, refIDs.RefSettingIDs, requireActive)
	if err != nil {
		return apperrors.New(err)
	}
	err = uc.checkRefAppsExistence(ctx, db, refIDs.RefAppIDs, requireActive)
	if err != nil {
		return apperrors.New(err)
	}
	err = uc.checkRefUsersExistence(ctx, db, refIDs.RefUserIDs, requireActive)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (uc *BaseUC) checkRefSettingsExistence(
	ctx context.Context,
	db database.IDB,
	req *BaseSettingReq,
	refSettingIDs []string,
	requireActive bool,
) (err error) {
	if len(refSettingIDs) == 0 {
		return nil
	}
	settings, _, err := uc.SettingRepo.List(ctx, db, req.Scope, nil,
		bunex.SelectWhere("setting.id IN (?)", bunex.List(refSettingIDs)),
		bunex.SelectWhereIf(requireActive, "setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return apperrors.New(err)
	}
	for _, refSettingID := range refSettingIDs {
		found := entityutil.FindByID(settings, refSettingID)
		if found == nil {
			return apperrors.NewNotFound("Setting").WithMsgLog("setting %s not found", refSettingID)
		}
	}
	return nil
}

func (uc *BaseUC) checkRefAppsExistence(
	ctx context.Context,
	db database.IDB,
	refAppIDs []string,
	requireActive bool,
) (err error) {
	if len(refAppIDs) == 0 {
		return nil
	}
	_, err = uc.AppService.LoadApps(ctx, db, "", refAppIDs, requireActive, requireActive,
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (uc *BaseUC) checkRefUsersExistence(
	ctx context.Context,
	db database.IDB,
	refUserIDs []string,
	requireActive bool,
) (err error) {
	if len(refUserIDs) == 0 {
		return nil
	}
	_, err = uc.UserService.LoadUsers(ctx, db, refUserIDs, requireActive)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (uc *BaseUC) ensureSettingDefaultUniqueness(
	ctx context.Context,
	db database.IDB,
	req *BaseSettingReq,
	setting *entity.Setting,
) error {
	err := uc.SettingRepo.UpdateClearDefaultFlag(ctx, db, req.Scope, req.Type, &setting.Kind, setting.ID,
		bunex.UpdateWithDeleted(),
	)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func TransformSettingBase(setting *entity.Setting) (resp *BaseSettingResp, err error) {
	if setting == nil {
		return nil, nil
	}
	if err = copier.Copy(&resp, setting); err != nil {
		return nil, apperrors.New(err)
	}
	if setting.ObjectID != setting.CurrentObjectID {
		resp.Inherited = true
	}
	return resp, nil
}

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
	Type  base.SettingType   `json:"-" mapstructure:"-"`
	Scope *base.SettingScope `json:"-" mapstructure:"-"`
}

type BaseSettingResp struct {
	ID              string             `json:"id"`
	Type            base.SettingType   `json:"type"`
	Name            string             `json:"name"`
	Kind            string             `json:"kind,omitempty"`
	Status          base.SettingStatus `json:"status"`
	Inherited       bool               `json:"inherited,omitempty"`
	AvailInProjects bool               `json:"availableInProjects,omitempty"`
	Default         bool               `json:"default,omitempty"`
	UpdateVer       int                `json:"updateVer"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	ExpireAt  *time.Time `json:"expireAt,omitempty" copy:",nilonzero"`
}

func (uc *BaseUC) loadSettingByID(
	ctx context.Context,
	db database.IDB,
	req *BaseSettingReq,
	id string,
	requireActive bool,
	opts ...bunex.SelectQueryOption,
) (setting *entity.Setting, err error) {
	setting, err = uc.SettingRepo.GetByID(ctx, db, req.Scope, req.Type, id, requireActive, opts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
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
		return apperrors.Wrap(err)
	}
	if setting != nil {
		return apperrors.NewAlreadyExist(strutil.ToPascalCase(string(req.Type))).
			WithMsgLog("%s '%s' already exists", req.Type, setting.Name)
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
		bunex.SelectWhere("setting.id IN (?)", bunex.In(refSettingIDs)),
		bunex.SelectWhereIf(requireActive, "setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	for _, refSettingID := range refSettingIDs {
		found := entityutil.FindByID(settings, refSettingID)
		if found == nil {
			return apperrors.NewNotFound("Setting").WithMsgLog("setting %s not found", refSettingID)
		}
	}
	return nil
}

func (uc *BaseUC) ensureSettingDefaultUniqueness(
	ctx context.Context,
	db database.IDB,
	req *BaseSettingReq,
	setting *entity.Setting,
) error {
	err := uc.SettingRepo.UpdateClearDefaultFlag(ctx, db, req.Scope, req.Type, setting.ID,
		bunex.UpdateWithDeleted(),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func TransformSettingBase(setting *entity.Setting) (resp *BaseSettingResp, err error) {
	if setting == nil {
		return nil, nil
	}
	if err = copier.Copy(&resp, setting); err != nil {
		return nil, apperrors.Wrap(err)
	}
	if setting.ObjectID != setting.CurrentObjectID {
		resp.Inherited = true
	}
	return resp, nil
}

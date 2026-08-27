package settinginitserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

const (
	appPlacementSettingName = "App placement settings"
)

func (s *service) initDefaultAppPlacementSettings(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	appPlacementSetting := &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       base.ObjectScopeGlobal,
		Type:        base.SettingTypeAppPlacement,
		Status:      base.SettingStatusActive,
		Name:        appPlacementSettingName,
		Inheritable: true,
		Default:     true,
		Version:     entity.CurrentAppPlacementVersion,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}
	appPlacement := &entity.AppPlacementSettings{
		ExcludeBuildNodes:   true,
		ExcludeManagerNodes: true,
	}
	appPlacementSetting.MustSetData(appPlacement)

	err = s.settingRepo.Insert(ctx, db, appPlacementSetting)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

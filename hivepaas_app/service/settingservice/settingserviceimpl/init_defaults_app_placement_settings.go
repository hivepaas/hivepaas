package settingserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
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
		ID:              gofn.Must(ulid.NewStringULID()),
		Scope:           base.ObjectScopeGlobal,
		Type:            base.SettingTypeAppPlacementSettings,
		Status:          base.SettingStatusActive,
		Name:            appPlacementSettingName,
		AvailInProjects: true,
		Default:         true,
		Version:         entity.CurrentAppPlacementSettingsVersion,
		CreatedAt:       timeNow,
		UpdatedAt:       timeNow,
	}
	appPlacement := &entity.AppPlacementSettings{
		ExcludeBuildNodes:   true,
		ExcludeManagerNodes: true,
	}
	appPlacementSetting.MustSetData(appPlacement)

	err = s.settingRepo.Insert(ctx, db, appPlacementSetting)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

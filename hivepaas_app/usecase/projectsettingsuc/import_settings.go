package projectsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc/projectsettingsdto"
)

func (uc *UC) ImportSettingsToProject(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectsettingsdto.ImportSettingsToProjectReq,
) (*projectsettingsdto.ImportSettingsToProjectResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &settingImportData{}
		err := uc.loadSettingsForImport(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingSettingImportData{}
		uc.preparePersistingSettingImports(req, data, persistingData)

		err = uc.persistSettingImports(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectsettingsdto.ImportSettingsToProjectResp{}, nil
}

type settingImportData struct {
	Project  *entity.Project
	Settings []*entity.Setting
}

type persistingSettingImportData struct {
	SharedSettings []*entity.SharedSetting
}

func (uc *UC) loadSettingsForImport(
	ctx context.Context,
	db database.Tx,
	req *projectsettingsdto.ImportSettingsToProjectReq,
	data *settingImportData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ProjectID,
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF project"),
		bunex.SelectRelation("ProjectEnvs"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Project = project

	settingIDs := req.Settings.ToIDStringSlice()
	settings, err := uc.settingRepo.ListByIDs(ctx, db, nil, settingIDs, false)
	if err != nil {
		return apperrors.Wrap(err)
	}
	for _, setting := range settings {
		// Ignore if the setting belongs to the project already
		if setting.ObjectID != project.ID {
			data.Settings = append(data.Settings, setting)
		}
	}

	settingMap := entityutil.SliceToIDMap(settings)
	for _, id := range settingIDs {
		if _, exists := settingMap[id]; !exists {
			return apperrors.Wrap(apperrors.ErrSettingNotFound).WithParam("Name", id)
		}
	}

	return nil
}

func (uc *UC) preparePersistingSettingImports(
	req *projectsettingsdto.ImportSettingsToProjectReq,
	data *settingImportData,
	persistingData *persistingSettingImportData,
) {
	timeNow := timeutil.NowUTC()
	for _, setting := range data.Settings {
		sharedSetting := &entity.SharedSetting{
			Scope:       base.ObjectScopeProject,
			ObjectID:    data.Project.ID,
			SettingID:   setting.ID,
			CanViewData: req.CanViewData,
			CreatedAt:   timeNow,
		}
		persistingData.SharedSettings = append(persistingData.SharedSettings, sharedSetting)
	}
}

func (uc *UC) persistSettingImports(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingSettingImportData,
) error {
	err := uc.sharedSettingRepo.UpsertMulti(ctx, db, persistingData.SharedSettings,
		entity.SharedSettingUpsertingConflictCols, entity.SharedSettingUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

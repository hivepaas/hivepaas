package projectuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/entityutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectuc/projectdto"
)

func (uc *UC) ImportSettingsToProject(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.ImportSettingsToProjectReq,
) (*projectdto.ImportSettingsToProjectResp, error) {
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

	return &projectdto.ImportSettingsToProjectResp{}, nil
}

type settingImportData struct {
	Project  *entity.Project
	Settings []*entity.Setting
}

type persistingSettingImportData struct {
	ProjectSharedSettings []*entity.ProjectSharedSetting
}

func (uc *UC) loadSettingsForImport(
	ctx context.Context,
	db database.Tx,
	req *projectdto.ImportSettingsToProjectReq,
	data *settingImportData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ProjectID,
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF project"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Project = project

	settingIDs := req.Settings.ToIDStringSlice()
	settings, err := uc.settingRepo.ListByIDs(ctx, db, base.NewSettingScopeGlobal(), settingIDs, false)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Settings = settings

	settingMap := entityutil.SliceToIDMap(settings)
	for _, id := range settingIDs {
		if _, exists := settingMap[id]; !exists {
			return apperrors.NewNotFound("Setting").
				WithMsgLog("setting %s not found", id)
		}
	}

	return nil
}

func (uc *UC) preparePersistingSettingImports(
	req *projectdto.ImportSettingsToProjectReq,
	data *settingImportData,
	persistingData *persistingSettingImportData,
) {
	timeNow := timeutil.NowUTC()
	for _, setting := range data.Settings {
		persistingData.ProjectSharedSettings = append(persistingData.ProjectSharedSettings,
			&entity.ProjectSharedSetting{
				ProjectID:       data.Project.ID,
				SettingID:       setting.ID,
				DataViewAllowed: req.DataViewAllowed,
				CreatedAt:       timeNow,
			})
	}
}

func (uc *UC) persistSettingImports(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingSettingImportData,
) error {
	err := uc.projectSharedSettingRepo.UpsertMulti(ctx, db, persistingData.ProjectSharedSettings,
		entity.ProjectSharedSettingUpsertingConflictCols, entity.ProjectSharedSettingUpsertingUpdateCols)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

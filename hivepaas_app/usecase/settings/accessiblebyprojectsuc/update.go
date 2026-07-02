package accessiblebyprojectsuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/accessiblebyprojectsuc/accessiblebyprojectsdto"
)

func (uc *UC) UpdateAccessibleByProjects(
	ctx context.Context,
	auth *basedto.Auth,
	req *accessiblebyprojectsdto.UpdateAccessibleByProjectsReq,
) (*accessiblebyprojectsdto.UpdateAccessibleByProjectsResp, error) {
	err := transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		data := &updateAccessibleByProjectsData{}
		err := uc.loadAccessibleByProjectsData(ctx, db, req, data)
		if err != nil {
			return apperrors.New(err)
		}
		if !data.HasChanges {
			return nil
		}

		persistingData := &persistingAccessibleByProjectsData{}
		uc.preparePersistingAccessibleByProjectsUpdate(req, data, persistingData)

		return uc.persistAccessibleByProjectsData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &accessiblebyprojectsdto.UpdateAccessibleByProjectsResp{}, nil
}

type updateAccessibleByProjectsData struct {
	AddingAccesses   []*accessiblebyprojectsdto.AccessibleByProjectReq
	DeletingAccesses []*entity.ProjectSharedSetting
	HasChanges       bool
}

type persistingAccessibleByProjectsData struct {
	UpsertingSharedSettings []*entity.ProjectSharedSetting
}

func (uc *UC) loadAccessibleByProjectsData(
	ctx context.Context,
	db database.IDB,
	req *accessiblebyprojectsdto.UpdateAccessibleByProjectsReq,
	data *updateAccessibleByProjectsData,
) error {
	setting, err := uc.SettingRepo.GetByID(ctx, db, nil, "", req.SettingID, false,
		bunex.SelectRelation("AccessibleByProjects"),
		bunex.SelectFor("UPDATE OF setting"),
	)
	if err != nil {
		return apperrors.New(err)
	}

	currProjectIDMap := make(map[string]struct{}, len(setting.AccessibleByProjects))
	for _, ap := range setting.AccessibleByProjects {
		currProjectIDMap[ap.ProjectID] = struct{}{}
	}
	newProjectIDMap := make(map[string]struct{}, len(req.AccessibleByProjects))
	for _, ap := range req.AccessibleByProjects {
		newProjectIDMap[ap.ID] = struct{}{}
	}

	// Make sure all projects exist
	_, err = uc.ProjectService.LoadProjects(ctx, db, gofn.MapKeys(newProjectIDMap), true)
	if err != nil {
		return apperrors.New(err)
	}

	// Projects in the current list, but not in the updating list, will be removed
	for _, ap := range setting.AccessibleByProjects {
		if _, exist := newProjectIDMap[ap.ProjectID]; !exist {
			data.DeletingAccesses = append(data.DeletingAccesses, ap)
		}
	}
	// Projects in the updating list, but not in the current list, will be added
	for _, ap := range req.AccessibleByProjects {
		if _, exist := currProjectIDMap[ap.ID]; !exist {
			data.AddingAccesses = append(data.AddingAccesses, ap)
		}
	}

	data.HasChanges = len(data.AddingAccesses) > 0 || len(data.DeletingAccesses) > 0
	return nil
}

func (uc *UC) preparePersistingAccessibleByProjectsUpdate(
	req *accessiblebyprojectsdto.UpdateAccessibleByProjectsReq,
	data *updateAccessibleByProjectsData,
	persistingData *persistingAccessibleByProjectsData,
) {
	timeNow := timeutil.NowUTC()

	for _, ap := range data.DeletingAccesses {
		ap.DeletedAt = timeNow
		persistingData.UpsertingSharedSettings = append(persistingData.UpsertingSharedSettings, ap)
	}
	for _, ap := range data.AddingAccesses {
		persistingData.UpsertingSharedSettings = append(persistingData.UpsertingSharedSettings,
			&entity.ProjectSharedSetting{
				SettingID: req.SettingID,
				ProjectID: ap.ID,
			})
	}
}

func (uc *UC) persistAccessibleByProjectsData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingAccessibleByProjectsData,
) error {
	err := uc.ProjectSharedSettingRepo.UpsertMulti(ctx, db, persistingData.UpsertingSharedSettings,
		entity.ProjectSharedSettingUpsertingConflictCols, entity.ProjectSharedSettingUpsertingUpdateCols)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

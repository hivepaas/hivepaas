package projectsettingsuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc/projectsettingsdto"
)

func (uc *UC) UpdateUserAccesses(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectsettingsdto.UpdateUserAccessesReq,
) (*projectsettingsdto.UpdateUserAccessesResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateUserAccessesData{}
		err := uc.loadUserAccessesForUpdate(ctx, db, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		persistingData := &persistingProjectData{}
		uc.prepareUpdatingUserAccesses(req, data, persistingData)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &projectsettingsdto.UpdateUserAccessesResp{}, nil
}

type updateUserAccessesData struct {
	Project         *entity.Project
	CurrentAccesses map[string]*entity.ACLPermission
}

func (uc *UC) loadUserAccessesForUpdate(
	ctx context.Context,
	db database.Tx,
	req *projectsettingsdto.UpdateUserAccessesReq,
	data *updateUserAccessesData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ProjectID,
		bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF project"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if project.UpdateVer != req.UpdateVer {
		return hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}
	data.Project = project

	// Loads all current accesses
	currAccesses, err := uc.permissionManager.LoadProjectRawAccesses(ctx, db, project.ID, nil)
	if err != nil {
		return hperrors.Wrap(err)
	}

	data.CurrentAccesses = make(map[string]*entity.ACLPermission, len(currAccesses))
	for _, access := range currAccesses {
		data.CurrentAccesses[access.SubjectID+"/"+access.ResourceID] = access
	}

	return nil
}

func (uc *UC) prepareUpdatingUserAccesses(
	req *projectsettingsdto.UpdateUserAccessesReq,
	data *updateUserAccessesData,
	persistingData *persistingProjectData,
) {
	project := data.Project
	currAccesses := data.CurrentAccesses
	timeNow := timeutil.NowUTC()
	mapUpsertingACLs := make(map[string]*entity.ACLPermission, len(currAccesses))

	for _, envAccesses := range req.EnvUserAccesses {
		projectEnvID := projecthelper.CalcProjectEnvID(project.ID, envAccesses.Name)
		for _, accessReq := range envAccesses.UserAccesses {
			key := accessReq.ID + "/" + projectEnvID
			currAccess := currAccesses[key]
			if currAccess == nil {
				currAccess = &entity.ACLPermission{
					SubjectType:  base.SubjectTypeUser,
					SubjectID:    accessReq.ID,
					ResourceType: base.ResourceTypeProjectEnv,
					ResourceID:   projectEnvID,
					Actions:      accessReq.Access,
					CreatedAt:    timeNow,
					UpdatedAt:    timeNow,
				}
				mapUpsertingACLs[key] = currAccess
			} else {
				delete(currAccesses, key)
				if !currAccess.Actions.Equal(accessReq.Access) {
					currAccess.Actions = accessReq.Access
					currAccess.UpdatedAt = timeNow
					mapUpsertingACLs[key] = currAccess
				}
			}
		}
	}
	upsertingACLs := gofn.MapValues(mapUpsertingACLs)
	// Remaining items in the current list need to delete
	for _, access := range currAccesses {
		access.DeletedAt = timeNow
		upsertingACLs = append(upsertingACLs, access)
	}

	if len(upsertingACLs) == 0 {
		return
	}
	persistingData.UpsertingACLPermissions = append(persistingData.UpsertingACLPermissions, upsertingACLs...)

	project.UpdatedAt = timeNow
	project.UpdateVer++
	persistingData.UpsertingProjects = append(persistingData.UpsertingProjects, project)
}

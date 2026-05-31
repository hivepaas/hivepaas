package projectsettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectsettingsuc/projectsettingsdto"
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
			return apperrors.Wrap(err)
		}

		persistingData := &persistingProjectData{}
		uc.prepareUpdatingUserAccesses(req, data, persistingData)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectsettingsdto.UpdateUserAccessesResp{}, nil
}

type updateUserAccessesData struct {
	Project *entity.Project
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
		bunex.SelectRelation("Accesses",
			bunex.SelectWhere("acl_permission.subj_type = ?", base.SubjectTypeUser),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Project = project

	return nil
}

func (uc *UC) prepareUpdatingUserAccesses(
	req *projectsettingsdto.UpdateUserAccessesReq,
	data *updateUserAccessesData,
	persistingData *persistingProjectData,
) {
	project := data.Project
	timeNow := timeutil.NowUTC()

	newAccessesByUserID := make(map[string]*projectsettingsdto.UserAccessReq)
	for _, accessReq := range req.UserAccesses {
		newAccessesByUserID[accessReq.ID] = accessReq
	}

	// Accesses to delete
	for _, access := range project.Accesses {
		if _, exists := newAccessesByUserID[access.SubjectID]; !exists {
			access.DeletedAt = timeNow
			persistingData.UpsertingACLPermissions = append(persistingData.UpsertingACLPermissions, access)
		}
	}
	// Accesses to update or insert
	for _, accessReq := range newAccessesByUserID {
		persistingData.UpsertingACLPermissions =
			append(persistingData.UpsertingACLPermissions, &entity.ACLPermission{
				SubjectType:  base.SubjectTypeUser,
				SubjectID:    accessReq.ID,
				ResourceType: base.ResourceTypeProject,
				ResourceID:   project.ID,
				Actions:      accessReq.Access,
				CreatedAt:    timeNow,
				UpdatedAt:    timeNow,
			})
	}
}

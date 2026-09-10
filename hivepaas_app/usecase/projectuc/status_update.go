package projectuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

func (uc *UC) UpdateProjectStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.UpdateProjectStatusReq,
) (*projectdto.UpdateProjectStatusResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		projectData := &updateProjectData{}
		err := uc.loadProjectDataForUpdateStatus(ctx, db, req, projectData)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if !projectData.HasChanges {
			return nil
		}

		before := projectData.Project.Status

		persistingData := &persistingProjectData{}
		uc.preparePersistingProjectStatusUpdate(req, projectData, persistingData)

		project := projectData.Project
		for _, env := range project.ProjectEnvs {
			env.Project = project
			// Run updates in a separate transaction to reduce lock time
			err = uc.projectService.ExecuteEnvInTx(ctx, env, true, func(db database.Tx) error {
				err := uc.projectService.SetProjectEnvStatus(ctx, db, env, project.Status, true)
				if err != nil {
					return hperrors.Wrap(err)
				}
				return nil
			})
			if err != nil {
				return hperrors.Wrap(err)
			}
		}

		if err := uc.persistData(ctx, db, persistingData); err != nil {
			return hperrors.Wrap(err)
		}

		// The environments were already switched above, each in its own
		// transaction to keep the lock short - so this entry stands for a change
		// that has partly landed outside the transaction it is written in.
		return uc.recordProjectWrite(ctx, db, auth, project,
			base.AuditLogTypeProjectUpdate, base.AuditLogSourceAPIUpdate, "status", auditdetail.New().
				Compare("status", before, req.Status).
				Set("envCount", len(project.ProjectEnvs)))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &projectdto.UpdateProjectStatusResp{}, nil
}

func (uc *UC) loadProjectDataForUpdateStatus(
	ctx context.Context,
	db database.IDB,
	req *projectdto.UpdateProjectStatusReq,
	data *updateProjectData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ID,
		bunex.SelectFor("UPDATE OF project"),
		bunex.SelectRelation("ProjectEnvs.Apps"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if project.UpdateVer != req.UpdateVer {
		return hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}
	data.Project = project
	data.HasChanges = project.Status != req.Status

	return nil
}

func (uc *UC) preparePersistingProjectStatusUpdate(
	req *projectdto.UpdateProjectStatusReq,
	data *updateProjectData,
	persistingData *persistingProjectData,
) {
	timeNow := timeutil.NowUTC()
	project := data.Project
	project.UpdateVer++
	project.Status = req.Status
	project.UpdatedAt = timeNow

	persistingData.UpsertingProjects = append(persistingData.UpsertingProjects, project)
}

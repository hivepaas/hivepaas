package projectuc

import (
	"context"
	"errors"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

func (uc *UC) UpdateProject(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.UpdateProjectReq,
) (*projectdto.UpdateProjectResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		projectData := &updateProjectData{}
		err := uc.loadProjectDataForUpdate(ctx, db, auth, req, projectData)
		if err != nil {
			return apperrors.New(err)
		}
		if !projectData.HasChanges {
			return nil
		}

		persistingData := &persistingProjectData{}
		uc.preparePersistingProjectUpdate(req, projectData, persistingData)

		return uc.persistData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &projectdto.UpdateProjectResp{}, nil
}

type updateProjectData struct {
	Project    *entity.Project
	HasChanges bool
}

func (uc *UC) loadProjectDataForUpdate(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *projectdto.UpdateProjectReq,
	data *updateProjectData,
) error {
	project, err := uc.projectRepo.GetByID(ctx, db, req.ID,
		bunex.SelectFor("UPDATE"),
		bunex.SelectRelation("Settings",
			// NOTE: now we only need to update Envs, expand this list if need more
			bunex.SelectWhere("setting.type = ?", base.SettingTypeProjectEnvs),
		),
	)
	if err != nil {
		return apperrors.New(err)
	}
	if project.UpdateVer != req.UpdateVer {
		return apperrors.New(apperrors.ErrUpdateVerMismatched)
	}
	data.Project = project

	// To update project status, use a separate API, so we don't update it
	req.Status = project.Status

	// If name changes, need to verify it uniqueness
	if !strings.EqualFold(req.Name, project.Name) {
		conflictProject, err := uc.projectRepo.GetByName(ctx, db, req.Name, bunex.SelectColumns("id"))
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.New(err)
		}
		if conflictProject != nil {
			return apperrors.NewAlreadyExist("Project").
				WithMsgLog("project name '%s' already exists", req.Name)
		}
	}

	// Only admin, current owner, and users have permission on Project module can change project owner
	if req.Owner.ID != project.OwnerID && !auth.User.IsAdmin() && auth.User.ID != project.OwnerID {
		hasPerm, err := uc.permissionManager.CheckAccess(ctx, db, auth, &permission.AccessCheck{
			SubjectType:    base.SubjectTypeUser,
			SubjectID:      auth.User.ID,
			ResourceModule: base.ResourceModuleProject,
			Action:         base.ActionTypeWrite,
		})
		if err != nil {
			return apperrors.New(err)
		}
		if !hasPerm {
			return apperrors.New(apperrors.ErrUnauthorized)
		}
	}

	// Validate project owner
	if req.Owner.ID != "" && req.Owner.ID != project.OwnerID {
		_, err = uc.userService.LoadUser(ctx, db, req.Owner.ID)
		if err != nil {
			return apperrors.New(err)
		}
	}

	data.HasChanges = true
	return nil
}

func (uc *UC) preparePersistingProjectUpdate(
	req *projectdto.UpdateProjectReq,
	data *updateProjectData,
	persistingData *persistingProjectData,
) {
	project := data.Project
	project.UpdateVer++
	timeNow := timeutil.NowUTC()

	uc.preparePersistingProjectBase(project, req.ProjectBaseReq, timeNow, persistingData)
	uc.preparePersistingProjectEnvs(project, req.Envs, timeNow, persistingData)
	persistingData.ProjectsToDeleteTags = append(persistingData.ProjectsToDeleteTags, project.ID)
	uc.preparePersistingProjectTags(project, req.Tags, 0, persistingData)
}

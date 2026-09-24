package projectuc

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

func (uc *UC) CreateProject(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.CreateProjectReq,
) (_ *projectdto.CreateProjectResp, err error) {
	projectData := &createProjectData{}
	err = uc.loadProjectData(ctx, uc.db, auth, req, projectData)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	persistingData := &persistingProjectData{}
	err = uc.preparePersistingProject(ctx, req, projectData, persistingData)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Nothing outside the transaction was created by preparePersistingProject -
		// the default volume is just another setting now - so a failure here rolls
		// back cleanly with everything else and there is nothing left to clean up.
		project := persistingData.UpsertingProjects[0]
		return uc.recordProjectWrite(ctx, db, auth, project,
			base.AuditLogTypeProjectCreate, base.AuditLogSourceAPICreate, "create", auditdetail.New().
				Set("key", project.Key).
				Set("ownerId", project.OwnerID).
				Set("envCount", len(persistingData.UpsertingProjectEnvs)))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &projectdto.CreateProjectResp{
		Data: &basedto.ObjectIDResp{ID: persistingData.UpsertingProjects[0].ID},
	}, nil
}

type createProjectData struct {
	ProjectKey string
}

func (uc *UC) loadProjectData(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	req *projectdto.CreateProjectReq,
	data *createProjectData,
) error {
	data.ProjectKey = projecthelper.CalcProjectKey(req.Name)
	if gofn.Contain(base.UnallowedProjectKeys, data.ProjectKey) {
		return hperrors.Wrap(hperrors.ErrProjectNameNotAllowed).WithParam("Name", req.Name)
	}

	// Project key must be unique
	conflictProject, err := uc.projectRepo.GetByKey(ctx, db, data.ProjectKey, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if conflictProject != nil {
		return hperrors.NewAlreadyExist("Project").
			WithMsgLog("project key '%s' already exists", data.ProjectKey)
	}

	// Project name must be unique
	conflictProject, err = uc.projectRepo.GetByName(ctx, db, req.Name, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if conflictProject != nil {
		return hperrors.NewAlreadyExist("Project").
			WithMsgLog("project name '%s' already exists", req.Name)
	}

	// Validate project owner
	if req.Owner.ID != "" {
		_, err = uc.userService.LoadUser(ctx, db, req.Owner.ID, true)
		if err != nil {
			return hperrors.Wrap(err)
		}
	} else {
		req.Owner.ID = auth.User.ID
	}

	return nil
}

func (uc *UC) preparePersistingProject(
	ctx context.Context,
	req *projectdto.CreateProjectReq,
	data *createProjectData,
	persistingData *persistingProjectData,
) error {
	project := &entity.Project{
		ID:      gofn.Must(ulid.NewStringULID()),
		Key:     data.ProjectKey,
		Name:    req.Name,
		Status:  req.Status,
		Note:    req.Note,
		OwnerID: req.Owner.ID,
	}
	envs := make([]*projectservice.NewProjectEnvReq, 0, len(req.Envs))
	for _, env := range req.Envs {
		envs = append(envs, &projectservice.NewProjectEnvReq{Name: env.Name, Color: env.Color})
	}
	err := uc.projectService.PrepareNewProject(ctx, &projectservice.NewProjectReq{
		Project: project,
		Envs:    envs,
		Tags:    req.Tags,
		TimeNow: timeutil.NowUTC(),
	}, &persistingData.PersistingProjectData)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (uc *UC) preparePersistingProjectBase(
	project *entity.Project,
	req *projectdto.ProjectBaseReq,
	timeNow time.Time,
	persistingData *persistingProjectData,
) {
	project.Name = req.Name
	project.Status = req.Status
	project.Note = req.Note
	project.OwnerID = req.Owner.ID
	project.UpdatedAt = timeNow

	persistingData.UpsertingProjects = append(persistingData.UpsertingProjects, project)
}

func (uc *UC) preparePersistingProjectTags(
	project *entity.Project,
	tags []string,
	startIndex int,
	persistingData *persistingProjectData,
) {
	persistingData.UpsertingTags = append(persistingData.UpsertingTags,
		projectservice.NewProjectTags(project, tags, startIndex)...)
}

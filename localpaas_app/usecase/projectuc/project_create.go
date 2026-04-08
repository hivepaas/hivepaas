package projectuc

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/slugify"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/service/projectservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectuc/projectdto"
)

const (
	projectKeyMaxLen = 100

	projectWebhookName      = "default"
	projectWebhookSecretLen = 24
)

var (
	unallowedProjectKey = []string{"localpaas"}
)

func (uc *UC) CreateProject(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.CreateProjectReq,
) (*projectdto.CreateProjectResp, error) {
	projectData := &createProjectData{}
	err := uc.loadProjectData(ctx, uc.db, auth, req, projectData)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	persistingData := &persistingProjectData{}
	uc.preparePersistingProject(req, projectData, persistingData)

	createdProject := persistingData.UpsertingProjects[0]
	networkCreated := false

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		// Create default network for the project
		_, err = uc.networkService.CreateProjectNetwork(ctx, createdProject)
		if err != nil {
			return apperrors.Wrap(err)
		}
		networkCreated = true

		return nil
	})
	if err != nil {
		// If network created, now need to remove it
		if networkCreated {
			_ = uc.networkService.RemoveProjectNetwork(ctx, createdProject)
		}
		return nil, apperrors.Wrap(err)
	}

	return &projectdto.CreateProjectResp{
		Data: &basedto.ObjectIDResp{ID: createdProject.ID},
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
	data.ProjectKey = slugify.SlugifyEx(req.Name, nil, projectKeyMaxLen)
	if gofn.Contain(unallowedProjectKey, data.ProjectKey) {
		return apperrors.New(apperrors.ErrNameUnavailable).WithMsgLog("project name is not allowed")
	}

	// Project key must be unique
	conflictProject, err := uc.projectRepo.GetByKey(ctx, db, data.ProjectKey, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if conflictProject != nil {
		return apperrors.NewAlreadyExist("Project").
			WithMsgLog("project key '%s' already exists", data.ProjectKey)
	}

	// Project name must be unique
	conflictProject, err = uc.projectRepo.GetByName(ctx, db, req.Name, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if conflictProject != nil {
		return apperrors.NewAlreadyExist("Project").
			WithMsgLog("project name '%s' already exists", req.Name)
	}

	// Validate project owner
	if req.Owner.ID != "" {
		_, err = uc.userService.LoadUser(ctx, db, req.Owner.ID)
		if err != nil {
			return apperrors.Wrap(err)
		}
	} else {
		req.Owner.ID = auth.User.ID
	}

	return nil
}

type persistingProjectData struct {
	projectservice.PersistingProjectData
}

func (uc *UC) preparePersistingProject(
	req *projectdto.CreateProjectReq,
	data *createProjectData,
	persistingData *persistingProjectData,
) {
	timeNow := timeutil.NowUTC()
	// Upserting project
	project := &entity.Project{
		ID:        gofn.Must(ulid.NewStringULID()),
		Key:       data.ProjectKey,
		CreatedAt: timeNow,
	}

	uc.preparePersistingProjectBase(project, req.ProjectBaseReq, timeNow, persistingData)
	uc.preparePersistingProjectTags(project, req.Tags, 0, persistingData)
	uc.preparePersistingProjectWebhook(project, timeNow, persistingData)
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
	startDisplayOrder int,
	persistingData *persistingProjectData,
) {
	displayOrder := startDisplayOrder
	for _, tag := range tags {
		persistingData.UpsertingTags = append(persistingData.UpsertingTags,
			&entity.ProjectTag{
				ProjectID:    project.ID,
				Tag:          tag,
				DisplayOrder: displayOrder,
			})
		displayOrder++
	}
}

func (uc *UC) preparePersistingProjectWebhook(
	project *entity.Project,
	timeNow time.Time,
	persistingData *persistingProjectData,
) {
	setting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.SettingScopeProject,
		ObjectID:  project.ID,
		Type:      base.SettingTypeRepoWebhook,
		Status:    base.SettingStatusActive,
		Name:      projectWebhookName,
		Default:   true,
		Version:   entity.CurrentRepoWebhookVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	setting.MustSetData(&entity.RepoWebhook{
		Secret: gofn.RandTokenAsHex(projectWebhookSecretLen),
	})
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)
}

func (uc *UC) persistData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingProjectData,
) error {
	err := uc.projectService.PersistProjectData(ctx, db, &persistingData.PersistingProjectData)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

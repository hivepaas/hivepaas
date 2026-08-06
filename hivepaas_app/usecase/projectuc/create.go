package projectuc

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

const (
	projectWebhookName      = "default"
	projectWebhookSecretLen = 24

	projectNotificationName = "default"
)

func (uc *UC) CreateProject(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.CreateProjectReq,
) (_ *projectdto.CreateProjectResp, err error) {
	projectData := &createProjectData{}
	err = uc.loadProjectData(ctx, uc.db, auth, req, projectData)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	persistingData := &persistingProjectData{}
	err = uc.preparePersistingProject(ctx, req, projectData, persistingData)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	defer func() {
		if err != nil || recover() != nil {
			_ = uc.cleanupOnFail(ctx, projectData)
		}
	}()

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectdto.CreateProjectResp{
		Data: &basedto.ObjectIDResp{ID: persistingData.UpsertingProjects[0].ID},
	}, nil
}

type createProjectData struct {
	ProjectKey string

	CreatedVolume *client.VolumeCreateResult
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
		return apperrors.Wrap(apperrors.ErrProjectNameNotAllowed).WithParam("Name", req.Name)
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
		_, err = uc.userService.LoadUser(ctx, db, req.Owner.ID, true)
		if err != nil {
			return apperrors.Wrap(err)
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
) (err error) {
	timeNow := timeutil.NowUTC()
	// Upserting project
	project := &entity.Project{
		ID:        gofn.Must(ulid.NewStringULID()),
		Key:       data.ProjectKey,
		CreatedAt: timeNow,
	}

	uc.preparePersistingProjectBase(project, req.ProjectBaseReq, timeNow, persistingData)
	uc.preparePersistingProjectEnvs(project, req.Envs, 0, timeNow, persistingData)
	uc.preparePersistingProjectTags(project, req.Tags, 0, persistingData)
	uc.preparePersistingProjectWebhook(project, timeNow, persistingData)
	uc.preparePersistingProjectNotificationDefault(project, timeNow, persistingData)
	err = uc.preparePersistingProjectDefaultVolume(ctx, project, data, persistingData)
	if err != nil {
		return apperrors.Wrap(err)
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

func (uc *UC) preparePersistingProjectEnvs(
	project *entity.Project,
	envs []*projectdto.ProjectEnvReq,
	startIndex int,
	timeNow time.Time,
	persistingData *persistingProjectData,
) {
	index := startIndex
	for _, env := range envs {
		persistingData.UpsertingProjectEnvs = append(persistingData.UpsertingProjectEnvs,
			&entity.ProjectEnv{
				ID:        projecthelper.CalcProjectEnvID(project.ID, env.Name),
				ProjectID: project.ID,
				Name:      env.Name,
				Key:       projecthelper.CalcProjectEnvKey(env.Name),
				Status:    base.ProjectStatusActive,
				Color:     env.Color,
				Index:     index,
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
			})
		index++
	}
}

func (uc *UC) preparePersistingProjectTags(
	project *entity.Project,
	tags []string,
	startIndex int,
	persistingData *persistingProjectData,
) {
	index := startIndex
	for _, tag := range tags {
		persistingData.UpsertingTags = append(persistingData.UpsertingTags,
			&entity.Tag{
				ObjectID: project.ID,
				Tag:      tag,
				Index:    index,
			})
		index++
	}
}

func (uc *UC) preparePersistingProjectWebhook(
	project *entity.Project,
	timeNow time.Time,
	persistingData *persistingProjectData,
) {
	setting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeProject,
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

func (uc *UC) preparePersistingProjectNotificationDefault(
	project *entity.Project,
	timeNow time.Time,
	persistingData *persistingProjectData,
) {
	setting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeProject,
		ObjectID:  project.ID,
		Type:      base.SettingTypeNotification,
		Status:    base.SettingStatusActive,
		Name:      projectNotificationName,
		Default:   true,
		Version:   entity.CurrentNotificationVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	setting.MustSetData(entity.NewNotificationDefaultForScope(entity.NewObjectScopeProject(project.ID)))
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)
}

func (uc *UC) preparePersistingProjectDefaultVolume(
	ctx context.Context,
	project *entity.Project,
	data *createProjectData,
	persistingData *persistingProjectData,
) error {
	setting, createRes, err := uc.volumeService.CreateProjectDefaultVolume(ctx, project)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.CreatedVolume = createRes
	persistingData.UpsertingSettings = append(persistingData.UpsertingSettings, setting)
	return nil
}

func (uc *UC) cleanupOnFail(
	ctx context.Context,
	data *createProjectData,
) (err error) {
	if data.CreatedVolume != nil {
		volID := dockerhelper.GetVolumeID(&data.CreatedVolume.Volume)
		_, e := uc.dockerManager.VolumeRemove(ctx, volID, true)
		if e != nil {
			err = errors.Join(err, e)
		}
	}
	return err
}

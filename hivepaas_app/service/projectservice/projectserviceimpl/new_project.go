package projectserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
)

const (
	projectWebhookSecretLen = 24
)

func (s *service) PrepareNewProject(
	ctx context.Context,
	req *projectservice.NewProjectReq,
	out *projectservice.PersistingProjectData,
) error {
	project, now := req.Project, req.TimeNow
	project.CreatedAt, project.UpdatedAt = now, now
	out.UpsertingProjects = append(out.UpsertingProjects, project)
	for i, env := range req.Envs {
		out.UpsertingProjectEnvs = append(out.UpsertingProjectEnvs,
			projectservice.NewProjectEnv(project, env.Name, env.Color, i, now))
	}
	out.UpsertingTags = append(out.UpsertingTags, projectservice.NewProjectTags(project, req.Tags, 0)...)
	out.UpsertingSettings = append(out.UpsertingSettings,
		newProjectWebhook(project, now), newProjectNotificationDefault(project, now))

	volume, err := s.volumeService.CreateProjectDefaultVolume(ctx, project)
	if err != nil {
		return hperrors.Wrap(err)
	}
	out.UpsertingSettings = append(out.UpsertingSettings, volume)
	return nil
}

func newProjectWebhook(project *entity.Project, now time.Time) *entity.Setting {
	setting := &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       base.ObjectScopeProject,
		ObjectID:    project.ID,
		Type:        base.SettingTypeRepoWebhook,
		Status:      base.SettingStatusActive,
		Name:        projectservice.NewProjectDefaults[base.SettingTypeRepoWebhook],
		Inheritable: true,
		Default:     true,
		Version:     entity.CurrentRepoWebhookVersion,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	// Kind is deliberately left unset: the provider is unknown when a project is
	// created, and an unset kind means the webhook accepts any of them - the
	// sender is identified per delivery, see webhookuc.detectWebhookKind.
	setting.MustSetData(&entity.RepoWebhook{
		Secret: entity.NewEncryptedField(gofn.RandTokenAsHex(projectWebhookSecretLen)),
	})
	return setting
}

func newProjectNotificationDefault(project *entity.Project, now time.Time) *entity.Setting {
	setting := &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       base.ObjectScopeProject,
		ObjectID:    project.ID,
		Type:        base.SettingTypeNotification,
		Status:      base.SettingStatusActive,
		Name:        projectservice.NewProjectDefaults[base.SettingTypeNotification],
		Inheritable: true,
		Default:     true,
		Version:     entity.CurrentNotificationVersion,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	setting.MustSetData(entity.NewNotificationDefaultForScope(entity.NewObjectScopeProject(project.ID)))
	return setting
}

func (s *service) CheckNewProjectName(ctx context.Context, db database.IDB, name string) (string, error) {
	key := projecthelper.CalcProjectKey(name)
	if gofn.Contain(base.UnallowedProjectKeys, key) {
		return "", hperrors.Wrap(hperrors.ErrProjectNameNotAllowed).WithParam("Name", name)
	}
	conflict, err := s.projectRepo.GetByKey(ctx, db, key, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return "", hperrors.Wrap(err)
	}
	if conflict != nil {
		return "", hperrors.NewAlreadyExist("Project").WithMsgLog("project key '%s' already exists", key)
	}
	conflict, err = s.projectRepo.GetByName(ctx, db, name, bunex.SelectColumns("id"))
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return "", hperrors.Wrap(err)
	}
	if conflict != nil {
		return "", hperrors.NewAlreadyExist("Project").WithMsgLog("project name '%s' already exists", name)
	}
	return key, nil
}

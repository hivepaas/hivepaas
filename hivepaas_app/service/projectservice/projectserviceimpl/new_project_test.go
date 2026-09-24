package projectserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

type fakeVolumeService struct {
	volumeservice.Service
}

func (fakeVolumeService) CreateProjectDefaultVolume(
	_ context.Context, project *entity.Project,
) (*entity.Setting, error) {
	return &entity.Setting{ID: "vol_" + project.ID, Type: base.SettingTypeClusterVolume}, nil
}

// A project is created with its envs and tags and the defaults every project
// gets, whoever creates it.
func TestPrepareNewProjectGivesTheDefaults(t *testing.T) {
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
	svc := &service{volumeService: fakeVolumeService{}}
	now := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)
	project := &entity.Project{ID: "p1", Key: "shop", Name: "Shop", OwnerID: "u1"}
	out := &projectservice.PersistingProjectData{}

	err = svc.PrepareNewProject(context.Background(), &projectservice.NewProjectReq{
		Project: project,
		Envs:    []*projectservice.NewProjectEnvReq{{Name: "development", Color: "blue"}, {Name: "prod"}},
		Tags:    []string{"team-a"},
		TimeNow: now,
	}, out)

	assert.NoError(t, err)
	assert.Equal(t, []*entity.Project{project}, out.UpsertingProjects)
	assert.Equal(t, now, project.CreatedAt)
	if assert.Len(t, out.UpsertingProjectEnvs, 2) {
		dev := out.UpsertingProjectEnvs[0]
		assert.Equal(t, []any{"p1:dev", "dev", "blue", 0}, []any{dev.ID, dev.Key, dev.Color, dev.Index})
		assert.Equal(t, 1, out.UpsertingProjectEnvs[1].Index)
	}
	assert.Equal(t, []*entity.Tag{{ObjectID: "p1", Tag: "team-a", Index: 0}}, out.UpsertingTags)

	types := make([]base.SettingType, 0, len(out.UpsertingSettings))
	for _, setting := range out.UpsertingSettings {
		types = append(types, setting.Type)
	}
	assert.Equal(t, []base.SettingType{
		base.SettingTypeRepoWebhook, base.SettingTypeNotification, base.SettingTypeClusterVolume,
	}, types)
	webhook, err := out.UpsertingSettings[0].AsRepoWebhook()
	if assert.NoError(t, err) {
		assert.False(t, webhook.Secret.IsEmpty(), "the webhook is created with a secret")
	}
}

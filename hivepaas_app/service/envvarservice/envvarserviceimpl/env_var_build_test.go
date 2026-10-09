package envvarserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

// noAppSettings has no routing, kind or Docker API settings for any app: the
// app publishes no variables of HivePaaS's of them.
type noAppSettings struct {
	repository.SettingRepo
}

func (noAppSettings) List(context.Context, database.IDB, *entity.ObjectScope, *basedto.Paging,
	...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	return nil, nil, nil
}

// An app's variables are built again with its previews': they inherit them,
// and a change saved to the app reaches them too.
func TestAnAppsVariablesAreBuiltWithItsPreviews(t *testing.T) {
	s := &service{settingRepo: noAppSettings{}}
	project := &entity.Project{ID: "p1", Name: "p1"}
	env := &entity.ProjectEnv{ID: "p1:dev", ProjectID: "p1", Project: project}
	app := &entity.App{ID: "web", Name: "web", ProjectID: "p1", ProjectEnvID: "p1:dev", Project: project, ProjectEnv: env}
	preview := &entity.App{ID: "web-pr", Name: "web-pr", ParentID: "web", ProjectID: "p1", ProjectEnvID: "p1:dev"}
	env.Apps = []*entity.App{app, preview}
	vars := []*envvarservice.EnvVar{{EnvVar: &entity.EnvVar{Key: "GREETING", Value: "hello"}}}

	result, err := s.BuildEnvVarsForAllAppsInApp(context.Background(), nil, &envvarservice.BuildEnvVarsInAppReq{
		App:                   app,
		DataLoadFunc:          envvarservice.NewStaticEnvLoadFunc(vars, nil),
		InheritedDataLoadFunc: envvarservice.NewStaticEnvLoadFunc(nil, nil),
	}, nil, false, false)

	assert.NoError(t, err)
	built := make([]string, 0, len(result))
	for _, data := range result {
		built = append(built, data.App.ID)
	}
	assert.Equal(t, []string{"web", "web-pr"}, built)
}

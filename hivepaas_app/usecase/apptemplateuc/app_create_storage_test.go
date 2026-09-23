package apptemplateuc

import (
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

func plannedApp(name string, doc *specmodel.AppDoc) *appToProvision {
	return &appToProvision{name: name, result: &templaterender.Result{Doc: doc}}
}

func docWithMounts(mounts map[string]specmodel.Mount, settings map[string]any) *specmodel.AppDoc {
	return &specmodel.AppDoc{
		Deployment: &specmodel.Deployment{Storage: &specmodel.Storage{Mounts: mounts}},
		Settings:   settings,
	}
}

func TestAppStorageQueriesAskAboutEveryVolumeTheAppMounts(t *testing.T) {
	app := plannedApp("Postgres", docWithMounts(map[string]specmodel.Mount{
		"/var/lib/postgresql": {Type: mount.TypeVolume, Source: "vol-1"},
		"/backups": {
			Type:          mount.TypeVolume,
			Source:        "vol-2",
			VolumeOptions: &specmodel.VolumeOptions{Subpath: "dumps"},
		},
	}, nil))

	queries := appStorageQueries(app, "postgres", &entity.Project{Key: "shop"}, "prod")

	assert.Len(t, queries, 2)
	for _, query := range queries {
		assert.Equal(t, "postgres", query.AppKey)
		assert.Equal(t, "postgres", query.App.Key)
		assert.Equal(t, "prod", query.App.ProjectEnv.Key)
		if query.VolumeID == "vol-2" {
			assert.Equal(t, "dumps", query.Subpath)
		}
	}
}

// A mount that names another app reaches somebody else's directory on purpose -
// a file manager over the database beside it. What is in there is not a leftover
// of this app, and warning about it would be warning about the feature working.
func TestAppStorageQueriesLeaveAnotherAppsDirectoryAlone(t *testing.T) {
	app := plannedApp("Files", docWithMounts(map[string]specmodel.Mount{
		"/data": {
			Type:      mount.TypeVolume,
			Source:    "vol-1",
			SourceApp: &specmodel.MountSourceApp{App: "postgres"},
		},
		"/config": {Type: mount.TypeVolume, Source: "vol-1"},
	}, nil))

	queries := appStorageQueries(app, "files", &entity.Project{Key: "shop"}, "prod")

	assert.Len(t, queries, 1)
	assert.Empty(t, queries[0].Subpath)
}

func TestAppStorageQueriesIgnoreWhatIsNotAVolume(t *testing.T) {
	app := plannedApp("Api", docWithMounts(map[string]specmodel.Mount{
		"/tmp":  {Type: mount.TypeTmpfs},
		"/host": {Type: mount.TypeBind, Source: "/srv/data"},
	}, nil))

	assert.Empty(t, appStorageQueries(app, "api", &entity.Project{Key: "shop"}, "prod"))
}

func TestAppStorageQueriesAreEmptyWithoutStorage(t *testing.T) {
	assert.Empty(t, appStorageQueries(plannedApp("Api", &specmodel.AppDoc{}), "api",
		&entity.Project{Key: "shop"}, "prod"))
}

// The distinction the warning is built on: a database started on somebody else's
// data keeps the password that data was created with.
func TestAppIsDatabaseReadsTheRenderedKind(t *testing.T) {
	database := plannedApp("Postgres", docWithMounts(nil, map[string]any{
		"kind": map[string]any{"category": "database", "engine": "postgres"},
	}))
	webapp := plannedApp("Grafana", docWithMounts(nil, map[string]any{
		"kind": map[string]any{"category": "webapp", "engine": "grafana"},
	}))

	assert.True(t, appIsDatabase(database))
	assert.False(t, appIsDatabase(webapp))
	assert.False(t, appIsDatabase(plannedApp("Api", &specmodel.AppDoc{})))
}

// A template that creates several apps is asked about all of them.
//
// A component is an app like any other - it just has no template of its own -
// and a dependency is created by the same request. Either being left out would
// mean the one app of a Supabase or a Grafana that actually holds a database
// goes unchecked, which is the only one that matters.
func TestEveryAppOfARequestIsAskedAbout(t *testing.T) {
	volumeMount := map[string]specmodel.Mount{
		"/data": {Type: mount.TypeVolume, Source: "vol-1"},
	}
	rendered := &apptemplateservice.RenderResp{
		Result: &templaterender.Result{Doc: docWithMounts(volumeMount, nil)},
		Dependencies: []*apptemplateservice.RenderedDependency{{
			Name:    "db",
			AppName: "Shop db",
			Render: &apptemplateservice.RenderResp{
				TemplateResp: apptemplateservice.TemplateResp{
					Template: &templatemodel.Template{
						Metadata: templatemodel.Metadata{Name: "postgres"},
					},
				},
				Result: &templaterender.Result{Doc: docWithMounts(volumeMount, nil)},
			},
		}},
		Components: []*apptemplateservice.RenderedComponent{
			{
				Name:    "worker",
				AppName: "Shop worker",
				Result:  &templaterender.Result{Doc: docWithMounts(volumeMount, nil)},
			},
			{
				Name:    "web",
				AppName: "Shop",
				Primary: true,
				Result:  &templaterender.Result{Doc: docWithMounts(volumeMount, nil)},
			},
		},
	}

	var asked []string
	for _, app := range planApps(&apptemplatedto.CreateAppFromTemplateReq{Name: "Shop"}, rendered) {
		key := app.key()
		for range appStorageQueries(app, key, &entity.Project{Key: "shop"}, "prod") {
			asked = append(asked, key)
		}
	}

	// The dependency, the component that is not primary, and the app the person
	// named - which carries the primary component's render.
	assert.ElementsMatch(t, []string{"shop-db", "shop-worker", "shop"}, asked)
}

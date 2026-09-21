package apptemplateuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// sharedMountTemplateYAML is a file manager: an app whose whole purpose is to
// work on the files of another.
const sharedMountTemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: files
  title: Files
  tagline: Test file manager
  description: Test.
  categories: [databases/sql]
  icon: icons/files.svg
  requires: {versionCode: v000001}
parameters:
  - {name: targetApp, title: App to manage, type: app}
  - {name: dataVolume, title: Data volume, type: volume}
versions:
  - {name: "2", release: "2.0", default: true, image: "filebrowser/filebrowser:v2.44.0"}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
    storage:
      mounts:
        /srv/data:
          type: volume
          source: "${{ params.dataVolume }}"
          sourceApp: {app: "${{ params.targetApp }}", write: true}
  settings:
    routing: {port: 80}
`

type sharedMountAppRepo struct {
	repository.AppRepo
	apps []*entity.App
}

func (f *sharedMountAppRepo) List(
	_ context.Context, _ database.IDB, _ string, _ *basedto.Paging, _ ...bunex.SelectQueryOption,
) ([]*entity.App, *basedto.PagingMeta, error) {
	return f.apps, nil, nil
}

func sharedMountApps(t *testing.T, targetApp string) []*appToProvision {
	t.Helper()
	tmpl, err := templatemodel.DecodeTemplate([]byte(sharedMountTemplateYAML))
	assert.NoError(t, err)
	result, err := templaterender.Render(&templaterender.Request{
		Template: tmpl,
		Params:   map[string]any{"targetApp": targetApp, "dataVolume": "vol-1"},
	})
	assert.NoError(t, err)
	return []*appToProvision{{
		id:   "app-1",
		name: "files",
		rendered: &apptemplateservice.RenderResp{
			TemplateResp: apptemplateservice.TemplateResp{Template: tmpl},
			Result:       result,
		},
	}}
}

func sharedMountReq() *apptemplatedto.CreateAppFromTemplateReq {
	return &apptemplatedto.CreateAppFromTemplateReq{ProjectID: "p1", ProjectEnvID: "p1:dev"}
}

// The grant is everything in that app's directory, so it is checked against that
// app rather than against the project.
func TestCheckSharedMountsNeedsWriteOnTheNamedApp(t *testing.T) {
	permissions := &fakePermissionManager{granted: false}
	owner := &entity.App{ID: "app-2", Name: "Postgres", Key: "postgres", ProjectID: "p1", ProjectEnvID: "p1:dev"}
	uc := &UC{permissionManager: permissions, appRepo: &sharedMountAppRepo{apps: []*entity.App{owner}}}

	err := uc.checkSharedMounts(context.Background(), testAuth(), sharedMountReq(), sharedMountApps(t, "postgres"))

	assert.ErrorIs(t, err, hperrors.ErrUnauthorized)
	var hpErr hperrors.HPError
	assert.ErrorAs(t, err, &hpErr)
	detail := hpErr.Build("en").Detail
	assert.Contains(t, detail, "read and change the files of postgres")
	assert.Contains(t, detail, "/srv/data")

	assert.Len(t, permissions.checked, 1)
	check, ok := permissions.checked[0].(*permission.AppAccessCheck)
	assert.True(t, ok)
	assert.Equal(t, owner.ID, check.AppID)
	assert.Equal(t, base.ActionTypeWrite, check.Action)
}

func TestCheckSharedMountsPassesWhenTheCallerMayHaveIt(t *testing.T) {
	owner := &entity.App{ID: "app-2", Name: "Postgres", Key: "postgres", ProjectID: "p1", ProjectEnvID: "p1:dev"}
	uc := &UC{
		permissionManager: &fakePermissionManager{granted: true},
		appRepo:           &sharedMountAppRepo{apps: []*entity.App{owner}},
	}

	assert.NoError(t, uc.checkSharedMounts(
		context.Background(), testAuth(), sharedMountReq(), sharedMountApps(t, "postgres")))
}

// A name with no app behind it would provision cleanly and then mount the app's
// own directory, which is not what the template asked for.
func TestCheckSharedMountsRefusesAnAppThatIsNotHere(t *testing.T) {
	uc := &UC{
		permissionManager: &fakePermissionManager{granted: true},
		appRepo:           &sharedMountAppRepo{},
	}

	err := uc.checkSharedMounts(
		context.Background(), testAuth(), sharedMountReq(), sharedMountApps(t, "nosuchapp"))

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
}

// An app of this very request is nobody else's: both are created by the same
// person in the same click, so there is nothing to ask.
func TestCheckSharedMountsExemptsAnAppOfTheSameRequest(t *testing.T) {
	permissions := &fakePermissionManager{granted: false}
	uc := &UC{permissionManager: permissions, appRepo: &sharedMountAppRepo{}}

	apps := sharedMountApps(t, "postgres")
	apps = append(apps, &appToProvision{id: "app-2", name: "postgres", role: "db",
		rendered: apps[0].rendered})

	assert.NoError(t, uc.checkSharedMounts(context.Background(), testAuth(), sharedMountReq(), apps))
	assert.Empty(t, permissions.checked)
}

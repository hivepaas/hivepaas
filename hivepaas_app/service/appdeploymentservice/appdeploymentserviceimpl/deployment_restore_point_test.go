package appdeploymentserviceimpl

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// markingManaged answers a mark as made, or with err, and keeps what it was asked.
type markingManaged struct {
	managedservice.Service
	asked []*managedservice.MarkPointReq
	err   error
}

func (f *markingManaged) MarkRestorePoint(_ context.Context, req *managedservice.MarkPointReq) (
	*managedservice.MarkedPoint, error) {
	f.asked = append(f.asked, req)
	if f.err != nil {
		return nil, f.err
	}
	return &managedservice.MarkedPoint{App: &entity.App{Name: "shop-db"},
		Point: &entity.ManagedRestorePoint{Name: req.Name, Position: "0/3000060"}}, nil
}

type namedApps struct {
	appservice.Service
}

func (namedApps) LoadApp(context.Context, database.IDB, string, string, bool, bool,
	...bunex.SelectQueryOption) (*entity.App, error) {
	return &entity.App{Name: "shop-db"}, nil
}

var deployMade = time.Date(2026, 10, 18, 9, 30, 15, 0, time.UTC)

func pointDeployment(database string) *appDeploymentData {
	deployment := &entity.Deployment{ID: "dep-1", CreatedAt: deployMade,
		Settings: &entity.AppDeploymentSettings{RestorePointBeforeDeploy: entity.ObjectID{ID: database}}}
	return &appDeploymentData{
		AppDeploymentReq: &appdeploymentservice.AppDeploymentReq{
			TaskExecData: &queue.TaskExecData{LogStore: tasklog.NewNullStore()},
		},
		App:        &entity.App{ID: "web-1", ProjectID: "p1", Key: "shop-web", Name: "shop web"},
		Deployment: deployment,
	}
}

// An app that names a database has a point marked there, a deployment's,
// named after the app and when the deployment was made.
func TestADeploymentMarksTheRestorePointItAsksFor(t *testing.T) {
	managed := &markingManaged{}
	s := &service{managedService: managed}

	err := s.deployStepRestorePoint(context.Background(), pointDeployment("db-1"))

	assert.NoError(t, err)
	if assert.Len(t, managed.asked, 1) {
		req := managed.asked[0]
		assert.Equal(t, "p1", req.ProjectID)
		assert.Equal(t, "db-1", req.AppID)
		assert.Equal(t, "deploy-shop-web-20261018093015", req.Name)
		assert.Equal(t, entity.ManagedPointOriginDeploy, req.Origin)
		assert.True(t, req.ExistingOK, "a deployment run again finds its point made")
		assert.Contains(t, req.Note, "dep-1")
		assert.Empty(t, req.By)
	}
}

func TestADeploymentThatAsksForNoPointMarksNone(t *testing.T) {
	managed := &markingManaged{}

	err := (&service{managedService: managed}).deployStepRestorePoint(context.Background(), pointDeployment(""))

	assert.NoError(t, err)
	assert.Empty(t, managed.asked)
}

// A point that cannot be made fails the deployment, naming the database and why.
func TestADeploymentWhosePointCannotBeMadeFails(t *testing.T) {
	managed := &markingManaged{err: hperrors.Wrap(hperrors.ErrAppManagedRestorePointNotArchived)}
	s := &service{managedService: managed, appService: namedApps{}}

	err := s.deployStepRestorePoint(context.Background(), pointDeployment("db-1"))

	assert.ErrorIs(t, err, hperrors.ErrAppDeployRestorePointFailed)
	info, _ := hperrors.ParseError(err, "en")
	assert.Contains(t, info.Detail, "shop-db")
	assert.Contains(t, info.Detail, "not archived in time")
	assert.NotContains(t, info.Detail, "ERR_APP_MANAGED_RESTORE_POINT_NOT_ARCHIVED", "the reason is said, not coded")
}

// The name fits PostgreSQL's 63 bytes whatever the key, and is the same for one
// deployment whenever it runs.
func TestADeploymentsPointName(t *testing.T) {
	assert.Equal(t, "deploy-web-20261018093015", deployPointName("web", deployMade))
	assert.Equal(t, deployPointName("web", deployMade), deployPointName("web", deployMade.In(time.Local)))

	assert.Equal(t, "deploy-"+strings.Repeat("a", 40)+"-20261018093015",
		deployPointName(strings.Repeat("a", 40)+"-"+strings.Repeat("b", 30), deployMade),
		"the key cut to fit, and a dash it ends on left off")
	assert.LessOrEqual(t, len(deployPointName(strings.Repeat("k", 100), deployMade)), 63)
}

package appmanageduc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/postgres"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

type postgresManaged struct {
	managedservice.Service
}

func (f *postgresManaged) Engine(name templatemodel.ManagedEngine) (managedmodel.Engine, bool) {
	return postgres.New(), name == templatemodel.ManagedEnginePostgres
}

type savedSettings struct {
	repository.SettingRepo
	saved []*entity.Setting
}

func (f *savedSettings) Update(_ context.Context, _ database.IDB, setting *entity.Setting,
	_ ...bunex.UpdateQueryOption) error {
	f.saved = append(f.saved, setting)
	return nil
}

func pointsApp(points ...*entity.ManagedRestorePoint) *entity.App {
	managed := setting("managed-1", base.SettingTypeAppManaged, &entity.AppManagedSettings{Engine: "postgres",
		Database: &entity.ManagedDatabase{RestorePoints: points}})
	return &entity.App{ID: "app-1", Settings: []*entity.Setting{managed}}
}

func pointsUC(app *entity.App) (*UC, *savedSettings) {
	settings := &savedSettings{}
	return &UC{appService: &fakeApps{app: app}, managedService: &postgresManaged{}, settingRepo: settings}, settings
}

func savedPoints(t *testing.T, settings *savedSettings) []*entity.ManagedRestorePoint {
	t.Helper()
	if !assert.Len(t, settings.saved, 1) {
		return nil
	}
	return settings.saved[0].MustAsAppManagedSettings().Database.RestorePoints
}

// A person's request reaches the engine's service as asked, with who asked:
// the service makes and records the point (its tests are its own).
func TestAPersonsPointRequest(t *testing.T) {
	keep := time.Date(2026, 11, 1, 12, 0, 0, 0, time.FixedZone("ICT", 7*3600))

	req := markPointReq(&basedto.Auth{User: &basedto.User{User: &entity.User{ID: "u1"}}},
		&appmanageddto.CreateRestorePointReq{ProjectID: "p1", AppID: "app-1", Name: "before-v42",
			Note: "migration 42", KeepUntil: &keep}, nil)

	assert.Equal(t, "p1", req.ProjectID)
	assert.Equal(t, "app-1", req.AppID)
	assert.Equal(t, "before-v42", req.Name)
	assert.Equal(t, "migration 42", req.Note)
	assert.Equal(t, "u1", req.By)
	assert.Equal(t, keep.UTC(), req.KeepUntil)
	assert.Empty(t, req.Origin, "a person's")
	assert.False(t, req.ExistingOK, "a name taken is refused")
	assert.Empty(t, markPointReq(nil, &appmanageddto.CreateRestorePointReq{Name: "x"}, nil).By)
}

func TestARestorePointsKeepAndNoteChange(t *testing.T) {
	point := &entity.ManagedRestorePoint{Name: "p", Position: "0/1", KeepUntil: time.Now().Add(time.Hour)}
	uc, settings := pointsUC(pointsApp(point))

	_, _, err := uc.updateRestorePoint(context.Background(), database.Tx{},
		&appmanageddto.UpdateRestorePointReq{Name: "p", Note: "kept no longer"})

	assert.NoError(t, err)
	if points := savedPoints(t, settings); assert.Len(t, points, 1) {
		assert.True(t, points[0].KeepUntil.IsZero(), "a null keepUntil keeps it no longer")
		assert.Equal(t, "kept no longer", points[0].Note)
	}
}

// Deleting a point removes its record. The archive is left as it is: the use
// case has nothing that runs in the database.
func TestDeletingARestorePointLeavesTheArchiveAlone(t *testing.T) {
	uc, settings := pointsUC(pointsApp(&entity.ManagedRestorePoint{Name: "a", Position: "0/1"},
		&entity.ManagedRestorePoint{Name: "b", Position: "0/2"}))

	_, _, err := uc.deleteRestorePoint(context.Background(), database.Tx{},
		&appmanageddto.DeleteRestorePointReq{Name: "a"})

	assert.NoError(t, err)
	if points := savedPoints(t, settings); assert.Len(t, points, 1) {
		assert.Equal(t, "b", points[0].Name)
	}

	_, _, err = uc.deleteRestorePoint(context.Background(), database.Tx{},
		&appmanageddto.DeleteRestorePointReq{Name: "never-made"})
	assert.ErrorIs(t, err, hperrors.ErrNotFound)
}

// A point's name is held to what its engine takes, before anything is marked;
// so is a restore's target point.
func TestAPointNameIsHeldToTheEnginesLimit(t *testing.T) {
	uc := &UC{appService: &fakeApps{app: managedApp()},
		managedService: &fakeManaged{engine: &fakeEngine{pointNameMax: 10}}}
	var invalid hperrors.ValidationErrors

	_, err := uc.CreateRestorePoint(context.Background(), nil, &appmanageddto.CreateRestorePointReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", AppID: "app-1", Name: "before-v42x"})
	assert.ErrorAs(t, err, &invalid)

	_, err = uc.RestoreManagedApp(context.Background(), nil, &appmanageddto.RestoreManagedReq{
		ProjectID: "p1", ProjectEnvID: "p1:prod", AppID: "app-1",
		Target: appmanageddto.ManagedRestoreTargetReq{Kind: entity.ManagedRestoreToPoint, Point: "before-v42x"}})
	assert.ErrorAs(t, err, &invalid)
}

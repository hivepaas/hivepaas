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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

type postgresManaged struct {
	managedservice.Service
}

func (f *postgresManaged) Engine(name templatemodel.ManagedEngine) (managedmodel.Engine, bool) {
	return postgres.New(), name == templatemodel.ManagedEnginePostgres
}

// markExec answers a mark with its exit code and what it printed.
type markExec struct {
	schedjobexecservice.Service
	exit     int
	stdout   string
	commands []string
}

func (f *markExec) RunCommand(_ context.Context, _ database.Tx, req *schedjobexecservice.RunCommandReq) (
	*schedjobexecservice.RunCommandResp, error) {
	f.commands = append(f.commands, req.Command.Command)
	_, _ = req.StdoutWriter.Write([]byte(f.stdout))
	exit := f.exit
	resp := &schedjobexecservice.RunCommandResp{ExitCode: &exit}
	if exit != 0 {
		return resp, hperrors.Wrap(hperrors.ErrActionFailed)
	}
	return resp, nil
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
		Postgres: &entity.ManagedPostgres{RestorePoints: points}})
	return &entity.App{ID: "app-1", Settings: []*entity.Setting{managed}}
}

func pointsUC(app *entity.App, exec *markExec) (*UC, *savedSettings) {
	settings := &savedSettings{}
	return &UC{appService: &fakeApps{app: app}, managedService: &postgresManaged{}, schedJobExecService: exec,
		settingRepo: settings}, settings
}

func savedPoints(t *testing.T, settings *savedSettings) []*entity.ManagedRestorePoint {
	t.Helper()
	if !assert.Len(t, settings.saved, 1) {
		return nil
	}
	return settings.saved[0].MustAsAppManagedSettings().Postgres.RestorePoints
}

// A point is recorded once the WAL it is in is archived, with the LSN the mark
// printed. No license is asked (this UC has none to ask).
func TestARestorePointIsRecordedWithItsLSN(t *testing.T) {
	exec := &markExec{stdout: "LSN=0/3000060\n"}
	uc, settings := pointsUC(pointsApp(), exec)
	keep := time.Now().UTC().Add(30 * 24 * time.Hour)
	before := time.Now().UTC()

	_, point, err := uc.createRestorePoint(context.Background(), database.Tx{},
		&basedto.Auth{User: &basedto.User{User: &entity.User{ID: "u1"}}},
		&appmanageddto.CreateRestorePointReq{Name: "before-v42", Note: "migration 42", KeepUntil: &keep})

	assert.NoError(t, err)
	assert.Equal(t, []string{"/opt/hivepaas/bin/walg-mark before-v42 120"}, exec.commands)
	if points := savedPoints(t, settings); assert.Len(t, points, 1) {
		assert.Equal(t, point, points[0])
		assert.Equal(t, "0/3000060", point.LSN)
		assert.Equal(t, "u1", point.By)
		assert.Equal(t, "migration 42", point.Note)
		assert.Equal(t, keep, point.KeepUntil)
		assert.False(t, point.At.Before(before))
	}
}

func TestARestorePointIsRefusedWhenItCannotBeListed(t *testing.T) {
	existing := &entity.ManagedRestorePoint{Name: "before-v42", LSN: "0/1"}
	for name, tc := range map[string]struct {
		app      *entity.App
		pointArg string
		exec     *markExec
		want     error
		marked   bool
	}{
		"a name taken":          {pointsApp(existing), "before-v42", &markExec{}, hperrors.ErrAlreadyExist, false},
		"a name HivePaaS keeps": {pointsApp(), "hp-test-1", &markExec{}, hperrors.ErrArgumentInvalid, false},
		"its WAL not archived in time": {pointsApp(), "late", &markExec{exit: managedmodel.MarkExitNotArchived},
			hperrors.ErrAppManagedRestorePointNotArchived, true},
		"no LSN printed": {pointsApp(), "quiet", &markExec{stdout: "done\n"}, hperrors.ErrInternal, true},
	} {
		uc, settings := pointsUC(tc.app, tc.exec)

		_, _, err := uc.createRestorePoint(context.Background(), database.Tx{}, nil,
			&appmanageddto.CreateRestorePointReq{Name: tc.pointArg})

		assert.ErrorIs(t, err, tc.want, name)
		assert.Empty(t, settings.saved, "%s: nothing is recorded", name)
		assert.Equal(t, tc.marked, len(tc.exec.commands) > 0, name)
	}
}

func TestARestorePointsKeepAndNoteChange(t *testing.T) {
	point := &entity.ManagedRestorePoint{Name: "p", LSN: "0/1", KeepUntil: time.Now().Add(time.Hour)}
	uc, settings := pointsUC(pointsApp(point), &markExec{})

	_, _, err := uc.updateRestorePoint(context.Background(), database.Tx{},
		&appmanageddto.UpdateRestorePointReq{Name: "p", Note: "kept no longer"})

	assert.NoError(t, err)
	if points := savedPoints(t, settings); assert.Len(t, points, 1) {
		assert.True(t, points[0].KeepUntil.IsZero(), "a null keepUntil keeps it no longer")
		assert.Equal(t, "kept no longer", points[0].Note)
	}
}

// Deleting a point removes its record, and runs nothing in the database: the
// archive is left as it is.
func TestDeletingARestorePointLeavesTheArchiveAlone(t *testing.T) {
	exec := &markExec{}
	uc, settings := pointsUC(pointsApp(&entity.ManagedRestorePoint{Name: "a", LSN: "0/1"},
		&entity.ManagedRestorePoint{Name: "b", LSN: "0/2"}), exec)

	_, _, err := uc.deleteRestorePoint(context.Background(), database.Tx{},
		&appmanageddto.DeleteRestorePointReq{Name: "a"})

	assert.NoError(t, err)
	if points := savedPoints(t, settings); assert.Len(t, points, 1) {
		assert.Equal(t, "b", points[0].Name)
	}
	assert.Empty(t, exec.commands)

	_, _, err = uc.deleteRestorePoint(context.Background(), database.Tx{},
		&appmanageddto.DeleteRestorePointReq{Name: "never-made"})
	assert.ErrorIs(t, err, hperrors.ErrNotFound)
}

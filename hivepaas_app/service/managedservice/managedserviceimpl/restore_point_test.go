package managedserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
)

type pointApps struct {
	appservice.Service
	app *entity.App
}

func (f *pointApps) LoadApp(context.Context, database.IDB, string, string, bool, bool,
	...bunex.SelectQueryOption) (*entity.App, error) {
	return f.app, nil
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
	managed := &entity.Setting{ID: "managed-1", Type: base.SettingTypeAppManaged, Status: base.SettingStatusActive}
	managed.MustSetData(&entity.AppManagedSettings{Engine: "postgres",
		Postgres: &entity.ManagedPostgres{RestorePoints: points}})
	return &entity.App{ID: "app-1", Settings: []*entity.Setting{managed}}
}

func pointsService(app *entity.App, exec *markExec) (*service, *savedSettings) {
	settings := &savedSettings{}
	return &service{registry: newRegistry(), settingRepo: settings, appService: &pointApps{app: app},
		schedJobExecService: exec}, settings
}

func savedPoints(t *testing.T, settings *savedSettings) []*entity.ManagedRestorePoint {
	t.Helper()
	if !assert.Len(t, settings.saved, 1) {
		return nil
	}
	return settings.saved[0].MustAsAppManagedSettings().Postgres.RestorePoints
}

// A point is recorded once the WAL it is in is archived, with the LSN the mark
// printed, and the caller's record runs once it is saved.
func TestARestorePointIsRecordedWithItsLSN(t *testing.T) {
	exec := &markExec{stdout: "LSN=0/3000060\n"}
	svc, settings := pointsService(pointsApp(), exec)
	keep := time.Now().UTC().Add(30 * 24 * time.Hour)
	before := time.Now().UTC()
	var recorded *entity.ManagedRestorePoint

	marked, err := svc.markRestorePoint(context.Background(), database.Tx{}, &managedservice.MarkPointReq{
		Name: "before-v42", Note: "migration 42", KeepUntil: keep, By: "u1",
		Record: func(_ context.Context, _ database.IDB, _ *entity.App, point *entity.ManagedRestorePoint) error {
			recorded = point
			return nil
		}})

	assert.NoError(t, err)
	assert.Equal(t, []string{"/opt/hivepaas/bin/walg-mark before-v42 120"}, exec.commands)
	if points := savedPoints(t, settings); assert.Len(t, points, 1) {
		point := points[0]
		assert.Equal(t, marked.Point, point)
		assert.Equal(t, point, recorded)
		assert.Equal(t, "0/3000060", point.LSN)
		assert.Equal(t, "u1", point.By)
		assert.Equal(t, "migration 42", point.Note)
		assert.Equal(t, keep, point.KeepUntil)
		assert.Empty(t, point.Origin, "a person's")
		assert.False(t, point.At.Before(before))
	}
	assert.False(t, marked.Existed)
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
		svc, settings := pointsService(tc.app, tc.exec)

		_, err := svc.markRestorePoint(context.Background(), database.Tx{},
			&managedservice.MarkPointReq{Name: tc.pointArg})

		assert.ErrorIs(t, err, tc.want, name)
		assert.Empty(t, settings.saved, "%s: nothing is recorded", name)
		assert.Equal(t, tc.marked, len(tc.exec.commands) > 0, name)
	}
}

// A deployment run again asks for the point it made the first time: it is
// answered as made, with nothing marked or saved.
func TestAnExistingPointCanCountAsMade(t *testing.T) {
	existing := &entity.ManagedRestorePoint{Name: "deploy-web-20261018120000", LSN: "0/1",
		Origin: entity.ManagedPointOriginDeploy}
	exec := &markExec{}
	svc, settings := pointsService(pointsApp(existing), exec)

	marked, err := svc.markRestorePoint(context.Background(), database.Tx{}, &managedservice.MarkPointReq{
		Name: existing.Name, Origin: entity.ManagedPointOriginDeploy, ExistingOK: true})

	assert.NoError(t, err)
	assert.True(t, marked.Existed)
	assert.Equal(t, existing.LSN, marked.Point.LSN)
	assert.Empty(t, exec.commands)
	assert.Empty(t, settings.saved)
}

// A deployment's point is recorded as one.
func TestADeploymentsPointIsRecordedAsOne(t *testing.T) {
	svc, settings := pointsService(pointsApp(), &markExec{stdout: "LSN=0/3000060\n"})

	_, err := svc.markRestorePoint(context.Background(), database.Tx{}, &managedservice.MarkPointReq{
		Name: "deploy-web-20261018120000", Origin: entity.ManagedPointOriginDeploy})

	assert.NoError(t, err)
	if points := savedPoints(t, settings); assert.Len(t, points, 1) {
		assert.Equal(t, entity.ManagedPointOriginDeploy, points[0].Origin)
		assert.Empty(t, points[0].By)
	}
}

// Pruning removes the deployments' points no base backup kept reaches, and
// saves only when it removes one.
func TestPruneDeployPoints(t *testing.T) {
	at := func(day int) time.Time { return time.Date(2026, 10, day, 12, 0, 0, 0, time.UTC) }
	old := &entity.ManagedRestorePoint{Name: "deploy-old", At: at(10), Origin: entity.ManagedPointOriginDeploy}
	fresh := &entity.ManagedRestorePoint{Name: "deploy-new", At: at(17), Origin: entity.ManagedPointOriginDeploy}
	person := &entity.ManagedRestorePoint{Name: "mine", At: at(10)}
	outputs := map[string]string{"BACKUPS": "base_000000010000000000000010@2026-10-16T02:00:00Z@100 " +
		"base_000000010000000000000020@2026-10-17T02:00:00Z@100"}
	svc, settings := pointsService(pointsApp(old, fresh, person), &markExec{})

	pruned, err := svc.PruneDeployPoints(context.Background(), nil, "p1", "app-1", outputs)

	assert.NoError(t, err)
	assert.Equal(t, []string{"deploy-old"}, pruned)
	if points := savedPoints(t, settings); assert.Len(t, points, 2) {
		assert.Equal(t, "deploy-new", points[0].Name)
		assert.Equal(t, "mine", points[1].Name, "a person's point stays")
	}

	svc, settings = pointsService(pointsApp(fresh, person), &markExec{})
	pruned, err = svc.PruneDeployPoints(context.Background(), nil, "p1", "app-1", outputs)
	assert.NoError(t, err)
	assert.Empty(t, pruned)
	assert.Empty(t, settings.saved, "nothing to prune, nothing saved")
}

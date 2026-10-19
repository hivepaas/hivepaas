package managedserviceimpl

import (
	"bytes"
	"context"
	"slices"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// markSlack is what a mark is given beyond its own wait: the exec's own start
// and end.
const markSlack = 30 * time.Second

// lockedManaged is a managed database loaded to change its engine's setting,
// its row held.
type lockedManaged struct {
	app     *entity.App
	setting *entity.Setting
	managed *entity.AppManagedSettings
	engine  managedmodel.Engine
	// writer is the cluster's primary the engine's commands run on: app, or the
	// member promoted.
	writer *entity.App
}

func (s *service) MarkRestorePoint(ctx context.Context, req *managedservice.MarkPointReq) (
	*managedservice.MarkedPoint, error) {
	var marked *managedservice.MarkedPoint
	err := transaction.Execute(ctx, s.db, func(db database.Tx) error {
		var txErr error
		marked, txErr = s.markRestorePoint(ctx, db, req)
		return txErr
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return marked, nil
}

func (s *service) markRestorePoint(ctx context.Context, db database.Tx, req *managedservice.MarkPointReq) (
	*managedservice.MarkedPoint, error) {
	locked, err := s.lockManaged(ctx, db, req.ProjectID, req.AppID)
	if err != nil {
		return nil, err
	}
	state := locked.managed.Database
	if existing := state.RestorePoint(req.Name); existing != nil {
		if req.ExistingOK {
			return &managedservice.MarkedPoint{App: locked.app, Point: existing, Existed: true}, nil
		}
		return nil, hperrors.NewAlreadyExist("Restore point").WithExtraDetail("%s", req.Name)
	}
	mark, err := locked.engine.MarkRestorePoint(locked.managed, req.Name)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	// When it was asked for: the point is after it, and the base backup a restore
	// to it fetches is the latest that starts before.
	at := timeutil.NowUTC()
	position, err := s.runMark(ctx, db, locked, mark)
	if err != nil {
		return nil, err
	}
	point := &entity.ManagedRestorePoint{Name: req.Name, Position: position, At: at, By: req.By,
		Note: req.Note, Origin: req.Origin}
	if !req.KeepUntil.IsZero() {
		point.KeepUntil = req.KeepUntil.UTC()
	}
	state.RestorePoints = append(state.RestorePoints, point)
	if err = s.saveManaged(ctx, db, locked); err != nil {
		return nil, err
	}
	if req.Record != nil {
		if err = req.Record(ctx, db, locked.app, point); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return &managedservice.MarkedPoint{App: locked.app, Point: point}, nil
}

// runMark runs the engine's mark in the database's task, and reads the point's
// position from what it printed.
func (s *service) runMark(ctx context.Context, db database.Tx, locked *lockedManaged,
	mark managedmodel.MarkCommand) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, mark.Timeout+markSlack)
	defer cancel()
	stdout := &bytes.Buffer{}
	resp, err := s.schedJobExecService.RunCommand(ctx, db, &schedjobexecservice.RunCommandReq{
		TaskExecData: &queue.TaskExecData{Task: &entity.Task{ID: gofn.Must(ulid.NewStringULID())},
			LogStore: tasklog.NewNullStore()},
		App:          locked.writer,
		Command:      &entity.CommandTemplate{Command: mark.Command},
		StdoutWriter: stdout,
	})
	if resp != nil && resp.ExitCode != nil && *resp.ExitCode == managedmodel.MarkExitNotArchived {
		return "", hperrors.Wrap(hperrors.ErrAppManagedRestorePointNotArchived)
	}
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	// The mark did not start: the database's service runs no container, as one
	// scaled to zero.
	if resp == nil || resp.ExitCode == nil {
		return "", hperrors.Wrap(hperrors.ErrActiveContainerNotFound).WithParam("App", locked.writer.Name)
	}
	if *resp.ExitCode != 0 {
		return "", hperrors.Wrap(hperrors.ErrActionFailed).
			WithExtraDetail("the restore point could not be made: the mark exited with %d", *resp.ExitCode)
	}
	position, err := locked.engine.MarkResult(stdout.String())
	return position, hperrors.Wrap(err)
}

func (s *service) PruneDeployPoints(ctx context.Context, db database.IDB, projectID, appID string,
	outputs map[string]string) ([]string, error) {
	locked, err := s.lockManaged(ctx, db, projectID, appID)
	if err != nil {
		return nil, err
	}
	stale := locked.engine.StaleDeployPoints(locked.managed, outputs)
	if len(stale) == 0 {
		return nil, nil
	}
	state := locked.managed.Database
	state.RestorePoints = slices.DeleteFunc(state.RestorePoints, func(p *entity.ManagedRestorePoint) bool {
		return slices.Contains(stale, p.Name)
	})
	return stale, s.saveManaged(ctx, db, locked)
}

// lockManaged loads a managed database with its engine's setting, its app's row
// held so that what changes the setting - a mark, a prune, an endpoint - takes
// turns. The project and its environment come along, which LoadApp checks are
// active.
func (s *service) lockManaged(ctx context.Context, db database.IDB, projectID, appID string) (
	*lockedManaged, error) {
	app, err := s.appService.LoadApp(ctx, db, projectID, appID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppManaged),
			bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	locked := &lockedManaged{app: app}
	for _, setting := range app.Settings {
		if setting.Type == base.SettingTypeAppManaged && setting.DeletedAt.IsZero() {
			locked.setting = setting
		}
	}
	if locked.setting == nil {
		return nil, hperrors.NewNotFound("Managed engine")
	}
	if locked.managed, err = locked.setting.AsAppManagedSettings(); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if locked.managed.Database == nil {
		return nil, hperrors.NewNotFound("Managed archive")
	}
	engine, found := s.Engine(templatemodel.ManagedEngine(locked.managed.Engine))
	if !found {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateIncompatible).
			WithExtraDetail("engine %s is not one this HivePaaS has", locked.managed.Engine)
	}
	locked.engine = engine
	if locked.writer, err = s.WriterOf(ctx, db, app, locked.managed); err != nil {
		return nil, err
	}
	return locked, nil
}

// WriterOf is the cluster's primary: the app, or the member promoted over it.
func (s *service) WriterOf(ctx context.Context, db database.IDB, app *entity.App,
	managed *entity.AppManagedSettings) (*entity.App, error) {
	if managed.Database == nil || managed.Database.Writer.ID == "" {
		return app, nil
	}
	writer, err := s.appService.LoadApp(ctx, db, app.ProjectID, managed.Database.Writer.ID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...))
	return writer, hperrors.Wrap(err)
}

// saveManaged stores the engine's setting as it was changed.
func (s *service) saveManaged(ctx context.Context, db database.IDB, locked *lockedManaged) error {
	if err := locked.setting.SetAppManagedData(locked.managed); err != nil {
		return hperrors.Wrap(err)
	}
	locked.setting.UpdateVer++
	locked.setting.UpdatedAt = timeutil.NowUTC()
	return hperrors.Wrap(s.settingRepo.Update(ctx, db, locked.setting,
		bunex.UpdateColumns("data", "version", "update_ver", "updated_at")))
}

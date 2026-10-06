package syscleanupserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

// errSystemAppsNeedAttention is a run that found what it leaves to a person:
// it fails the job, for its notification to say so.
var errSystemAppsNeedAttention = errors.New("system apps need attention")

// updateWithin is how recent a system update not yet done must be to count as
// running, as the update's own check has it.
const updateWithin = time.Hour

// systemAppsSync is one feature's sync: what it did, the deployments to
// schedule once committed, and the docker cleanup to run when it is rolled
// back.
type systemAppsSync struct {
	apps               []*entity.SystemAppSyncOutput
	obi                *entity.OBISyncOutput
	tasks              []*entity.Task
	cleanup            func(ctx context.Context) error
	obiSettingsChanged bool
}

// sysSyncSystemApps brings the apps HivePaaS runs for itself - the registry,
// the logging stack - to their settings, and checks OBI on the nodes. Each
// feature syncs in a savepoint of its own, so one that fails is rolled back
// alone - with what it made in docker - and the other goes on.
func (s *service) sysSyncSystemApps(ctx context.Context, db database.IDB, data *sysCleanupData) error {
	if data.SyncSystemApps == base.CleanupFlagFalse || !data.SysCleanupSettings.SystemAppsSyncEnabled() {
		return nil
	}
	out := &entity.SystemAppsSyncOutput{}
	data.TaskOutput.SystemApps = out

	updating, err := s.systemUpdating(ctx, db)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if updating {
		out.Skipped = "a system update is running"
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("System apps not synced: a system update is running",
			tasklog.TsNow))
		return nil
	}

	var errs []error
	for _, feature := range []struct {
		name string
		sync func(ctx context.Context, tx database.Tx) (*systemAppsSync, error)
	}{
		{"logging", s.syncLogging},
		{"registry", s.syncRegistry},
	} {
		synced, err := s.syncInSavepoint(ctx, db, data, feature.sync)
		if err != nil {
			errs = append(errs, fmt.Errorf("syncing the %s apps: %w", feature.name, err))
			_ = data.LogStore.Add(ctx, tasklog.NewErrFrame("Failed to sync the "+feature.name+" apps: "+
				err.Error(), tasklog.TsNow))
			out.Apps = append(out.Apps, &entity.SystemAppSyncOutput{Key: feature.name, Name: feature.name,
				Action: entity.SystemAppSyncFailed, Error: err.Error()})
			continue
		}
		out.Apps = append(out.Apps, synced.apps...)
		if synced.obi != nil {
			out.OBI = synced.obi
		}
	}

	if attention := needingAttention(out); len(attention) > 0 {
		errs = append(errs, fmt.Errorf("%w: %s", errSystemAppsNeedAttention, strings.Join(attention, "; ")))
	}
	return errors.Join(errs...)
}

// syncInSavepoint runs one feature's sync in a savepoint. Committed, it logs
// what it did, records a removal as a person's is, and schedules the
// deployments - and has the agents read OBI's settings again - once the task's
// transaction commits; rolled back, it takes down what it made in docker.
func (s *service) syncInSavepoint(
	ctx context.Context,
	db database.IDB,
	data *sysCleanupData,
	sync func(ctx context.Context, tx database.Tx) (*systemAppsSync, error),
) (*systemAppsSync, error) {
	var synced *systemAppsSync
	err := transaction.Execute(ctx, db, func(tx database.Tx) error {
		var e error
		synced, e = sync(ctx, tx)
		if e != nil {
			return e
		}
		return s.auditRemovals(ctx, tx, synced.apps)
	})
	if err != nil {
		if synced != nil && synced.cleanup != nil {
			err = errors.Join(err, synced.cleanup(context.WithoutCancel(ctx)))
		}
		return nil, hperrors.Wrap(err)
	}

	logSynced(ctx, data, synced)
	if len(synced.tasks) > 0 || synced.obiSettingsChanged {
		tasks, obiChanged := synced.tasks, synced.obiSettingsChanged
		data.OnPostTx(func() {
			ctx := context.WithoutCancel(ctx)
			if len(tasks) > 0 {
				if e := s.taskQueue.ScheduleTask(ctx, tasks...); e != nil {
					_ = data.LogStore.Add(ctx, tasklog.NewErrFrame("Failed to schedule the system apps' "+
						"deployments: "+e.Error(), tasklog.TsNow))
				}
			}
			if obiChanged {
				_ = s.obiSettings.Invalidate(ctx)
			}
		})
	}
	return synced, nil
}

func (s *service) syncLogging(ctx context.Context, tx database.Tx) (*systemAppsSync, error) {
	resp, err := s.loggingService.Sync(ctx, tx)
	if resp == nil {
		return nil, hperrors.Wrap(err)
	}
	return &systemAppsSync{apps: resp.Apps, obi: resp.OBI, tasks: resp.Tasks, cleanup: resp.Cleanup,
		obiSettingsChanged: resp.OBISettingsChanged}, hperrors.Wrap(err)
}

func (s *service) syncRegistry(ctx context.Context, tx database.Tx) (*systemAppsSync, error) {
	resp, err := s.registryService.Sync(ctx, tx)
	if resp == nil {
		return nil, hperrors.Wrap(err)
	}
	synced := &systemAppsSync{tasks: resp.Tasks, cleanup: resp.Cleanup}
	if resp.App != nil {
		synced.apps = []*entity.SystemAppSyncOutput{resp.App}
	}
	return synced, hperrors.Wrap(err)
}

// auditRemovals records each app the sync removed - one the settings no longer
// want, or one removed to be provisioned again - as the cleanup's removal of
// an orphaned app is recorded.
func (s *service) auditRemovals(ctx context.Context, tx database.Tx, apps []*entity.SystemAppSyncOutput) error {
	for _, app := range apps {
		if app.Action != entity.SystemAppSyncRemoved && app.Action != entity.SystemAppSyncRecreated {
			continue
		}
		removedID := app.AppID
		if app.PreviousAppID != "" {
			removedID = app.PreviousAppID
		}
		err := s.auditService.Record(ctx, tx, &auditservice.Entry{
			Type:     base.AuditLogTypeAppDelete,
			Scope:    base.ObjectScopeApp,
			ObjectID: removedID,
			Source:   base.AuditLogSourceSystemCleanup,
			Result:   base.AuditLogResultAllowed,
			ResType:  base.ResourceTypeApp,
			ResID:    removedID,
			ResName:  app.Name,
			Detail: auditdetail.New().
				Set("reason", "system app sync: "+app.Problem).
				Set("key", app.Key).
				Set("removeStorage", false).String(),
		})
		if err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

// systemUpdating says whether a system update is running: it moves these
// apps itself.
func (s *service) systemUpdating(ctx context.Context, db database.IDB) (bool, error) {
	tasks, _, err := s.taskRepo.ListByTarget(ctx, db, "", nil,
		bunex.SelectWhere("task.type = ?", base.TaskTypeSystemUpdate),
		bunex.SelectWhereIn("task.status IN (?)", base.TaskStatusNotStarted, base.TaskStatusInProgress),
		bunex.SelectWhere("task.created_at > ?", timeutil.NowUTC().Add(-updateWithin)),
		bunex.SelectLimit(1),
		bunex.SelectColumns("id"),
	)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return len(tasks) > 0, nil
}

// logSynced writes into the task's log what the sync did and found.
func logSynced(ctx context.Context, data *sysCleanupData, synced *systemAppsSync) {
	for _, app := range synced.apps {
		line := "System app " + app.Name + ": " + string(app.Action)
		if app.Problem != "" {
			line += " - " + app.Problem
		}
		frame := tasklog.NewOutFrame
		if app.Action == entity.SystemAppSyncReported {
			frame = tasklog.NewWarnFrame
		}
		_ = data.LogStore.Add(ctx, frame(line, tasklog.TsNow))
	}
	if synced.obi == nil {
		return
	}
	switch {
	case synced.obi.Problem != "":
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame("OBI: "+synced.obi.Problem, tasklog.TsNow))
	case synced.obi.Unknown != "":
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("OBI not checked: "+synced.obi.Unknown, tasklog.TsNow))
	}
	for _, node := range synced.obi.Nodes {
		line := "OBI on node " + nodeName(node) + ": " + string(node.Action)
		if node.Problem != "" {
			line += " - " + node.Problem
		}
		frame := tasklog.NewOutFrame
		if node.Action == entity.OBISyncReported {
			frame = tasklog.NewWarnFrame
		}
		_ = data.LogStore.Add(ctx, frame(line, tasklog.TsNow))
	}
}

// needingAttention is what the sync leaves to a person, one line each.
func needingAttention(out *entity.SystemAppsSyncOutput) []string {
	var lines []string
	for _, app := range out.Apps {
		if app.Action == entity.SystemAppSyncReported {
			lines = append(lines, app.Name+": "+app.Problem)
		}
	}
	if out.OBI == nil {
		return lines
	}
	if out.OBI.Problem != "" {
		lines = append(lines, "OBI: "+out.OBI.Problem)
	}
	for _, node := range out.OBI.Nodes {
		if node.Action == entity.OBISyncReported {
			lines = append(lines, "OBI on node "+nodeName(node)+": "+node.Problem)
		}
	}
	return lines
}

func nodeName(node *entity.OBINodeSyncOutput) string {
	if node.Hostname != "" {
		return node.Hostname
	}
	return node.NodeID
}

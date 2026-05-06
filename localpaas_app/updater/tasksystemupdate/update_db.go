package tasksystemupdate

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/fileutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	dbServiceUpdateCheckInterval     = time.Second * 5
	dbServiceRequiredRunningDuration = time.Second * 10
)

func (e *Executor) migrateDBSchema(
	ctx context.Context,
	data *taskData,
) (err error) {
	cfg := config.Current
	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, applog.NewOutFrame("Start migrating db schema...", applog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Migrating db schema finished in "+duration.String()+
				" with error: "+err.Error(), applog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Migrating db schema finished in "+duration.String(),
				applog.TsNow))
		}
	}()

	migBin, _ := fileutil.Lookup("sql-migrate", []string{
		"",
		"/usr/local/bin",
		"/usr/bin",
		"/localpaas",
	})
	if migBin == "" {
		return apperrors.NewNotFound("Binary 'sql-migrate'")
	}

	migConfigFile, _ := fileutil.Lookup("localpaas_app/db/dbconfig.yml", []string{
		"",
		"/localpaas",
	})
	if migConfigFile == "" {
		return apperrors.NewNotFound("Migration config file 'dbconfig.yml'")
	}

	cmd := exec.Command(migBin, "up", "-config="+migConfigFile, "-env=main")
	cmd.Env = []string{
		fmt.Sprintf("LP_DB_HOST=%v", cfg.DB.Host),
		fmt.Sprintf("LP_DB_PORT=%v", cfg.DB.Port),
		fmt.Sprintf("LP_DB_USER=%v", cfg.DB.User),
		fmt.Sprintf("LP_DB_PASSWORD=%v", cfg.DB.Password),
		fmt.Sprintf("LP_DB_DB_NAME=%v", cfg.DB.DBName),
	}

	res, err := cmd.CombinedOutput()
	for _, line := range strings.Split(reflectutil.UnsafeBytesToStr(res), "\n") {
		_ = data.LogStore.Add(ctx, applog.NewOutFrame(line, applog.TsNow))
	}

	return apperrors.Wrap(err)
}

func (e *Executor) migrateDBData(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) (err error) {
	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, applog.NewOutFrame("Start migrating db data...", applog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Migrating db data finished in "+duration.String()+
				" with error: "+err.Error(), applog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Migrating db data finished in "+duration.String(),
				applog.TsNow))
		}
	}()

	err = e.dbService.MigrateData(ctx, db)
	return apperrors.Wrap(err)
}

func (e *Executor) updateDbService(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) (err error) {
	args := data.UpdateArgs
	if args.TargetVersion.DbImage == "" {
		return nil
	}

	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating db service...", applog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating db service finished in "+duration.String()+
				" with error: "+err.Error(), applog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating db service finished in "+duration.String(),
				applog.TsNow))
		}
	}()

	dbSvc, err := e.lpAppService.GetLpDbSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}

	dbSvc.Spec.TaskTemplate.ContainerSpec.Image = args.TargetVersion.DbImage
	dbSvc.Spec.Mode.Replicated.Replicas = new(uint64(1))
	if dbSvc.Spec.UpdateConfig == nil {
		dbSvc.Spec.UpdateConfig = &swarm.UpdateConfig{}
	}
	dbSvc.Spec.UpdateConfig.FailureAction = swarm.UpdateFailureActionRollback
	dbSvc.Spec.UpdateConfig.MaxFailureRatio = 0.5

	_, err = e.dockerManager.ServiceUpdate(ctx, dbSvc.ID, &dbSvc.Version, &dbSvc.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Wait for the update to finish
	dbSvc, err = e.dockerManager.ServiceUpdateWait(ctx, dbSvc.ID, dbServiceUpdateCheckInterval)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if dbSvc.UpdateStatus != nil && dbSvc.UpdateStatus.State == swarm.UpdateStateRollbackCompleted {
		return apperrors.New(apperrors.ErrActionFailed).WithMsgLog("service db is rolled back")
	}

	// Wait for the service up and running
	running, err := e.dockerManager.ServiceWaitUntilRunning(ctx, dbSvc.ID, true,
		dbServiceRequiredRunningDuration, dbServiceUpdateCheckInterval)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if !running {
		return apperrors.New(apperrors.ErrServiceNotRunning).WithParam("Name", "db")
	}

	// Migrate DB schema
	err = e.migrateDBSchema(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Migrate DB data
	err = e.migrateDBData(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

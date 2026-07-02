package sysupdateserviceimpl

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	dbServiceUpdateCheckInterval     = time.Second * 5
	dbServiceRequiredRunningDuration = time.Second * 10
)

func (s *service) migrateDBSchema(
	ctx context.Context,
	data *sysUpdateData,
) (err error) {
	cfg := config.Current
	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Start migrating db schema...", tasklog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Migrating db schema finished in "+duration.String()+
				" with error: "+err.Error(), tasklog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Migrating db schema finished in "+duration.String(),
				tasklog.TsNow))
		}
	}()

	migBin, _ := fileutil.Lookup("sql-migrate", []string{
		"",
		"/usr/local/bin",
		"/usr/bin",
		"/hivepaas",
	})
	if migBin == "" {
		return apperrors.NewNotFound("BinObject 'sql-migrate'")
	}

	migConfigFile, _ := fileutil.Lookup("hivepaas_app/db/dbconfig.yml", []string{
		"",
		"/hivepaas",
	})
	if migConfigFile == "" {
		return apperrors.NewNotFound("Migration config file 'dbconfig.yml'")
	}

	cmd := exec.Command(migBin, "up", "-config="+migConfigFile, "-env=main")
	cmd.Env = []string{
		fmt.Sprintf("HP_DB_HOST=%v", cfg.DB.Host),
		fmt.Sprintf("HP_DB_PORT=%v", cfg.DB.Port),
		fmt.Sprintf("HP_DB_USER=%v", cfg.DB.User),
		fmt.Sprintf("HP_DB_PASSWORD=%v", cfg.DB.Password),
		fmt.Sprintf("HP_DB_DB_NAME=%v", cfg.DB.DBName),
	}

	res, err := cmd.CombinedOutput()
	for _, line := range strings.Split(reflectutil.UnsafeBytesToStr(res), "\n") {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(line, tasklog.TsNow))
	}

	return apperrors.New(err)
}

func (s *service) migrateDBData(
	ctx context.Context,
	db database.IDB,
	data *sysUpdateData,
) (err error) {
	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Start migrating db data...", tasklog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Migrating db data finished in "+duration.String()+
				" with error: "+err.Error(), tasklog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Migrating db data finished in "+duration.String(),
				tasklog.TsNow))
		}
	}()

	err = s.dbService.MigrateData(ctx, db)
	return apperrors.New(err)
}

func (s *service) updateDbService(
	ctx context.Context,
	db database.IDB,
	data *sysUpdateData,
) (err error) {
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())
	if args.TargetVersion.DbImage == "" {
		return nil
	}

	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Updating db service...", tasklog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Updating db service finished in "+duration.String()+
				" with error: "+err.Error(), tasklog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Updating db service finished in "+duration.String(),
				tasklog.TsNow))
		}
	}()

	dbSvc, err := s.hpAppService.GetHpDbSwarmService(ctx)
	if err != nil {
		return apperrors.New(err)
	}

	dbSvc.Spec.TaskTemplate.ContainerSpec.Image = args.TargetVersion.DbImage
	dbSvc.Spec.Mode.Replicated.Replicas = new(uint64(1))
	if dbSvc.Spec.UpdateConfig == nil {
		dbSvc.Spec.UpdateConfig = &swarm.UpdateConfig{}
	}
	dbSvc.Spec.UpdateConfig.FailureAction = swarm.UpdateFailureActionRollback
	dbSvc.Spec.UpdateConfig.MaxFailureRatio = 0.5

	_, err = s.dockerManager.ServiceUpdate(ctx, dbSvc.ID, &dbSvc.Version, &dbSvc.Spec)
	if err != nil {
		return apperrors.New(err)
	}

	// Wait for the update to finish
	dbSvc, err = s.dockerManager.ServiceUpdateWait(ctx, dbSvc.ID, dbServiceUpdateCheckInterval)
	if err != nil {
		return apperrors.New(err)
	}
	if dbSvc.UpdateStatus != nil && dbSvc.UpdateStatus.State == swarm.UpdateStateRollbackCompleted {
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame("service db is rolled back",
			tasklog.TsNow))
		return apperrors.New(apperrors.ErrActionFailed)
	}

	// Wait for the service up and running
	running, err := s.dockerManager.ServiceWaitUntilRunning(ctx, dbSvc.ID, true,
		dbServiceRequiredRunningDuration, dbServiceUpdateCheckInterval)
	if err != nil {
		return apperrors.New(err)
	}
	if !running {
		return apperrors.New(apperrors.ErrServiceNotRunning).WithParam("Name", "db")
	}

	// Migrate DB schema
	err = s.migrateDBSchema(ctx, data)
	if err != nil {
		return apperrors.New(err)
	}

	// Migrate DB data
	err = s.migrateDBData(ctx, db, data)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}

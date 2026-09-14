package sysupdateserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	// How long the database has to stay up before the migrations are run against
	// it, so a task that starts and immediately dies is not mistaken for ready.
	dbServiceRequiredRunningDuration = time.Second * 10
)

func (s *service) migrateDBSchema(
	ctx context.Context,
	data *sysUpdateData,
) (err error) {
	cfg := config.Current()
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
		return hperrors.NewNotFound("BinObject 'sql-migrate'")
	}

	migConfigFile, _ := fileutil.Lookup("hivepaas_app/db/dbconfig.yml", []string{
		"",
		"/hivepaas",
	})
	if migConfigFile == "" {
		return hperrors.NewNotFound("Migration config file 'dbconfig.yml'")
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

	return hperrors.Wrap(err)
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
	return hperrors.Wrap(err)
}

func (s *service) updateDbService(
	ctx context.Context,
	db database.IDB,
	data *sysUpdateData,
) (err error) {
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())

	upgrade, err := s.planDbMajorUpgrade(ctx, data, args.TargetVersion.DbImage, args.SkipBackup)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Moving postgres is optional; migrating is not. The schema and the data
	// migrations belong to the application's version, so a release that changes
	// them without changing the database image - or that names no database image
	// at all - still has to run them.
	err = s.updateServiceImage(ctx, data, serviceImageUpdate{
		What:        "db",
		Component:   base.HivepaasDbKey,
		TargetImage: args.TargetVersion.DbImage,
		Fetch: func(ctx context.Context) (*swarm.Service, error) {
			return s.hpAppService.GetHpDbSwarmService(ctx)
		},
		Mutate: func(spec *swarm.ServiceSpec) {
			spec.Mode.Replicated.Replicas = new(uint64(1))
			if upgrade != nil {
				upgrade.applyTo(spec)
			}
		},
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	// From here a major upgrade has an empty new cluster running and the old one
	// beside it, so anything that goes wrong is undone by putting the service
	// back rather than by restoring into the cluster being left behind.
	if upgrade != nil {
		defer func() {
			if err != nil {
				err = errors.Join(err, s.revertDbMajorUpgrade(ctx, data, upgrade))
			}
		}()
	}

	// undo answers a failure the way this particular update can. A major upgrade
	// has the deferred revert above; an ordinary one puts the database back from
	// the dump taken before the migrations ran.
	undo := func(cause error) error {
		if upgrade != nil {
			return cause
		}
		return errors.Join(cause, s.restoreDB(ctx, data))
	}

	dbSvc, err := s.hpAppService.GetHpDbSwarmService(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Checked even when the image did not move. The migrations are about to run
	// against it, and a database that is not accepting connections turns that
	// into a failed update with a half-applied schema.
	running, err := s.dockerManager.ServiceWaitUntilRunning(ctx, dbSvc.ID, true,
		dbServiceRequiredRunningDuration, serviceUpdateCheckInterval)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if !running {
		return hperrors.Wrap(hperrors.ErrServiceNotRunning).WithParam("Name", "db")
	}

	// A new cluster comes up empty. This is how the data gets into it - the one
	// place restoreDB is part of the plan rather than a recovery.
	if upgrade != nil {
		err = s.restoreDB(ctx, data)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	// The only failure the update undoes. Everything after this step leaves the
	// database migrated and working, and putting it back to answer a problem with
	// traefik would trade one inconsistency for another while discarding work.
	err = s.migrateDBSchema(ctx, data)
	if err != nil {
		return undo(hperrors.Wrap(err))
	}

	err = s.migrateDBData(ctx, db, data)
	if err != nil {
		return undo(hperrors.Wrap(err))
	}

	// Written last, once the new cluster has been loaded and migrated: before
	// that point the upgrade might still be put back, and pointing a later stack
	// deploy at a volume the database was never moved to would be worse than
	// pointing it at the old one.
	if upgrade != nil {
		err = s.recordDbVolume(ctx, data, upgrade)
		if err != nil {
			return undo(hperrors.Wrap(err))
		}
	}

	return nil
}

package tasksystemupdate

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/fileutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

func (e *Executor) migrateDBSchema(
	ctx context.Context,
	data *taskData,
) (err error) {
	cfg := config.Current
	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, applog.NewOutFrame("Start migrating DB schema...", applog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Migrating DB schema finished in "+duration.String()+
				" with error: "+err.Error(), applog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Migrating DB schema finished in "+duration.String(),
				applog.TsNow))
		}
	}()

	migBinDir := fileutil.Lookup("sql-migrate", []string{
		"/",
		"",
		"/usr/bin",
		"/usr/sbin",
		"/usr/local/bin",
		"/usr/local/sbin",
	})
	if migBinDir == "" {
		return apperrors.NewNotFound("Binary 'sql-migrate'")
	}
	migBin := filepath.Join(migBinDir, "sql-migrate")

	migConfigDir := fileutil.Lookup("dbconfig.yml", []string{
		"/localpaas_app/db",
		"localpaas_app/db",
	})
	if migConfigDir == "" {
		return apperrors.NewNotFound("Migration config file 'dbconfig.yml'")
	}
	migConfigFile := filepath.Join(migConfigDir, "dbconfig.yml")

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

package taskcronjobexec

import (
	"context"
	"reflect"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/jsonl"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	sysBackupSqlPageSize      = 1000
	sysBackupSqlPageSizeSmall = 100
)

var (
	sysBackupDBModels = []*sysBackupDBModel{
		{
			Type:   "db/user",
			Model:  new([]*entity.User),
			Orders: []string{"id"},
		},
		{
			Type:   "db/acl-permission",
			Model:  new([]*entity.ACLPermission),
			Orders: []string{"subject_id", "resource_id"},
		},
		{
			Type:         "db/login-trusted-device",
			Model:        new([]*entity.LoginTrustedDevice),
			Orders:       []string{"user_id", "device_id"},
			NoSoftDelete: true,
		},
		{
			Type:     "db/setting",
			Model:    new([]*entity.Setting),
			Orders:   []string{"id"},
			PageSize: sysBackupSqlPageSizeSmall, // some settings may have big data field
		},
		{
			Type:   "db/project",
			Model:  new([]*entity.Project),
			Orders: []string{"id"},
		},
		{
			Type:   "db/project-tag",
			Model:  new([]*entity.ProjectTag),
			Orders: []string{"project_id", "tag"},
		},
		{
			Type:   "db/project-shared-setting",
			Model:  new([]*entity.ProjectSharedSetting),
			Orders: []string{"project_id", "setting_id"},
		},
		{
			Type:   "db/app",
			Model:  new([]*entity.App),
			Orders: []string{"id"},
		},
		{
			Type:   "db/app-tag",
			Model:  new([]*entity.AppTag),
			Orders: []string{"app_id", "tag"},
		},
		{
			Type:   "db/deployment",
			Model:  new([]*entity.Deployment),
			Orders: []string{"id"},
		},
		{
			Type:   "db/task",
			Model:  new([]*entity.Task),
			Orders: []string{"id"},
		},
		{
			Type:         "db/task-log",
			Model:        new([]*entity.TaskLog),
			Orders:       []string{"id"},
			NoSoftDelete: true,
		},
		{
			Type:         "db/sys-error",
			Model:        new([]*entity.SysError),
			Orders:       []string{"id"},
			NoSoftDelete: true,
		},
	}
)

type sysBackupDBModel struct {
	Type         string
	Model        any
	PageSize     int
	Orders       []string
	NoSoftDelete bool
}

//nolint:gocognit
func (e *Executor) sysBackupDB(
	ctx context.Context,
	db database.IDB,
	sysBackup *entity.SystemBackup,
	jsonlW *jsonl.Writer,
	data *sysBackupTaskData,
) (err error) {
	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, applog.NewWarnFrame("Start backing up data from DB...", applog.TsNow))

	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewWarnFrame("DB backup finished in "+duration.String()+
				" with error: "+err.Error(), applog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("DB backup finished in "+duration.String(),
				applog.TsNow))
		}
	}()

	backupDeletedObjects := true
	if sysBackup.DBBackupConfig != nil {
		backupDeletedObjects = sysBackup.DBBackupConfig.BackupDeletedObjects
	}

	for _, model := range sysBackupDBModels {
		// NOTE: use cursor to speed up the SQL queries when possible
		offset := 0
		var lastID any
		for {
			pageSize := gofn.Coalesce(model.PageSize, sysBackupSqlPageSize)
			q := db.NewSelect().Model(model.Model).Limit(pageSize)
			if len(model.Orders) > 0 {
				q = q.Order(model.Orders...)
			}
			if lastID != nil {
				q = q.Where("id > ?", lastID) // Use cursor pagination
			} else {
				q = q.Offset(offset) // Use offset pagination (slower)
			}
			if backupDeletedObjects && !model.NoSoftDelete {
				q = q.WhereAllWithDeleted()
			}
			err = q.Scan(ctx)
			if err != nil {
				return apperrors.Wrap(err)
			}

			// Reflection to get the length of the slice
			val := reflect.ValueOf(model.Model).Elem()
			if val.Len() == 0 {
				break
			}

			err = jsonlW.WriteChunk(jsonl.NewChunk(model.Type, val.Interface()))
			if err != nil {
				return apperrors.Wrap(err)
			}

			if val.Len() < pageSize {
				break
			}
			offset += pageSize

			// Find last ID value for cursor pagination in the next query
			if len(model.Orders) == 1 && model.Orders[0] == "id" {
				lastItem := val.Index(val.Len() - 1)
				idStrEnt, _ := lastItem.Interface().(interface{ GetID() string })
				if idStrEnt != nil {
					lastID = idStrEnt.GetID()
					continue
				}
				idIntEnt, _ := lastItem.Interface().(interface{ GetID() int64 })
				if idIntEnt != nil {
					lastID = idIntEnt.GetID()
					continue
				}
			}
		}
	}

	return nil
}

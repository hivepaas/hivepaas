package systembackupuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

// auditSectionBackupRun is the section a backup started by hand is filed under.
//
// Prefixed with what it acts on, like the other sections of this type: it shares
// AuditLogTypeHivePaaSAction with the HivePaaS app's own restarts and with
// traefik's, and section is what separates them in the filter.
const auditSectionBackupRun = "system-backup-run"

// recordBackupRun records somebody starting a backup off-schedule.
//
// The settings behind it are recorded by the generic settings flow, which every
// write to a stored setting goes through. This is the other half: running one is
// not a write to the setting, so nothing in that flow sees it, and yet it is the
// moment the install's data is copied somewhere else. "Who took a copy, and when"
// is the question a backup leaves behind.
//
// An action rather than an update, and the resource is the backup setting it was
// started from - there is no backup row to point at until the job has run, and
// the task named in the detail is what leads to the outcome.
func (uc *UC) recordBackupRun(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	scope *entity.ObjectScope,
	backupSetting *entity.Setting,
	task *entity.Task,
) error {
	if scope == nil || backupSetting == nil || task == nil {
		return hperrors.NewArgumentInvalid("audited backup run")
	}

	err := auditservice.RecordAllowed(ctx, uc.AuditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeHivePaaSAction,
		Scope:    scope.ScopeType,
		ObjectID: scope.ScopeObjectID(),
		Source:   base.AuditLogSourceAPIAction,
		Section:  auditSectionBackupRun,
		Auth:     auth,
		ResType:  base.ResourceType(backupSetting.Type),
		ResID:    backupSetting.ID,
		ResName:  backupSetting.Name,
		Detail:   auditdetail.New().Set("taskId", task.ID).String(),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

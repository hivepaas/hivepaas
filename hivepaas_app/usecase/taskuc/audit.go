package taskuc

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

// recordTaskCancel records somebody stopping a task.
//
// Filed under the task's own scope, not the caller's. The two differ whenever a
// project's caller cancels an app's work, which the task scopes let them do -
// and it is the app's timeline that has to show the deploy being stopped, not
// only the project's. Filing it a level up would leave the app looking as though
// its deploy simply stopped on its own.
//
// A task under no object - a cleanup sweep, a system backup - carries the global
// or hivepaas scope and no object id, which is what those scopes mean.
//
// canceled tells the two outcomes apart: a task still waiting is marked canceled
// here and now, while one already running is only sent the request, over redis,
// and stops when it gets there. Both are somebody deciding the work should stop,
// which is what the entry is for; only one of them is finished when it is written.
func (uc *UC) recordTaskCancel(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	task *entity.Task,
	canceled bool,
) error {
	if task == nil || task.ID == "" {
		return hperrors.NewArgumentInvalid("audited task")
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeTaskCancel,
		Scope:    task.Scope,
		ObjectID: task.ObjectID,
		Source:   base.AuditLogSourceAPIAction,
		Auth:     auth,
		ResType:  base.ResourceTypeTask,
		ResID:    task.ID,
		// The type is what the entry is read for - a deploy stopped and a backup
		// stopped are not the same news - and it is free here, because the cancel
		// has the row in hand already.
		ResName: string(task.Type),
		// statusBefore is where the task was when somebody stopped it - waiting
		// its turn, or half way through. The task row says "canceled" from here
		// on and cannot answer that afterwards.
		Detail: auditdetail.New().
			Set("taskType", string(task.Type)).
			Set("statusBefore", string(task.Status)).
			Set("canceled", canceled).
			String(),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

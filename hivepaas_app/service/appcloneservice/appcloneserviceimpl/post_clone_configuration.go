package appcloneserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
)

// applyClonedConfiguration applies to docker what the copy's settings say.
//
// It keeps what was created in docker, because that is what cleanupOnFail has to
// remove: a rolled-back clone leaves no record of the secrets and configs it
// made, and they would otherwise sit there belonging to nothing.
func (s *service) applyClonedConfiguration(
	ctx context.Context,
	db database.IDB,
	data *appCloneData,
) error {
	resp, err := s.appProvisionService.ApplyAppConfiguration(ctx, db,
		&appprovisionservice.ApplyAppConfigurationReq{
			App:        data.DestApp,
			RefObjects: data.RefObjects,
		})
	if resp != nil {
		s.scheduleCertTasks(data, resp.CertTasks) //nolint:contextcheck // queued after this transaction
	}
	return hperrors.Wrap(err)
}

// scheduleCertTasks asks for the certificates the copy's domains have none for.
//
// The tasks' rows are written in the transaction this clone runs in, and a task
// can be claimed only once its row is there - so they are queued from the hook
// that runs after it commits. A clone driven from something that is not a task
// has no such hook, and leaves them to the queue's own scan, which is later
// rather than never.
func (s *service) scheduleCertTasks(data *appCloneData, tasks []*entity.Task) {
	if s.taskQueue == nil || len(tasks) == 0 || data.TaskExecData == nil {
		return
	}
	data.OnPostTx(func() {
		_ = s.taskQueue.ScheduleTask(context.Background(), tasks...)
	})
}

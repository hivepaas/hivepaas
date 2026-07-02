package schedjobservice

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

type Service interface {
	BuildCommandEnv(ctx context.Context, db database.IDB, app *entity.App, schedJob *entity.SchedJob) (
		res []*envvarservice.EnvVar, usedSecrets []*entity.Secret, err error)

	CreateSchedJobTask(job *entity.Setting, runAt, timeNow time.Time) (*entity.Task, error)
}

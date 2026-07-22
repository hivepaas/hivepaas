package schedjobservice

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

type Service interface {
	BuildCommandEnvVars(ctx context.Context, db database.IDB, app *entity.App, schedJob *entity.SchedJob) (
		[]*envvarservice.EnvVar, error)

	CreateSchedJobTask(job *entity.Setting, runAt, timeNow time.Time) (*entity.Task, error)
}

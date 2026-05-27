package schedjobservice

import (
	"context"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/envvarservice"
)

type Service interface {
	BuildCommandEnv(ctx context.Context, db database.IDB, app *entity.App, schedJob *entity.SchedJob) (
		res []*envvarservice.EnvVar, err error)

	CreateSchedJobTask(job *entity.Setting, runAt, timeNow time.Time) (*entity.Task, error)
}

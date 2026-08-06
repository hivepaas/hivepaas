package schedjobservice

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type Service interface {
	CreateSchedJobTask(job *entity.Setting, runAt, timeNow time.Time) (*entity.Task, error)
}

// Package getstarteduc serves the dashboard's Get started card: asking again
// for the dashboard's certificate, and closing the card.
package getstarteduc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type UC struct {
	db                *database.DB
	getStartedService getstartedservice.Service
	taskQueue         queue.TaskQueue
}

func New(
	db *database.DB,
	getStartedService getstartedservice.Service,
	taskQueue queue.TaskQueue,
) *UC {
	return &UC{
		db:                db,
		getStartedService: getStartedService,
		taskQueue:         taskQueue,
	}
}

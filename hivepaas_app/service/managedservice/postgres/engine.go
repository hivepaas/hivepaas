// Package postgres is the managed engine of Postgres: WAL-G archiving, base
// backups and their retention, restore tests, and restores to a point in time.
package postgres

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const (
	minAPI = 1
	maxAPI = 1
)

type Engine struct{}

func New() *Engine {
	return &Engine{}
}

func (e *Engine) Name() templatemodel.ManagedEngine {
	return templatemodel.ManagedEnginePostgres
}

func (e *Engine) APIRange() (int, int) {
	return minAPI, maxAPI
}

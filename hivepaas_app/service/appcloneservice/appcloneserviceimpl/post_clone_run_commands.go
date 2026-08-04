package appcloneserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func (s *service) runCommands(
	ctx context.Context,
	db database.IDB,
	data *appCloneData,
) (err error) {
	// Wait for the app to start
	return nil
}

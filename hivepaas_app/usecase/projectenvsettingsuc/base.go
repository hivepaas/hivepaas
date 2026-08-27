package projectenvsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
)

type persistingProjectEnvData struct {
	projectservice.PersistingProjectData
}

func (uc *UC) persistData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingProjectEnvData,
) error {
	err := uc.projectService.PersistProjectData(ctx, db, &persistingData.PersistingProjectData)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

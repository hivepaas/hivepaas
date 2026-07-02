package projectsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
)

type persistingProjectData struct {
	projectservice.PersistingProjectData
}

func (uc *UC) persistData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingProjectData,
) error {
	err := uc.projectService.PersistProjectData(ctx, db, &persistingData.PersistingProjectData)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

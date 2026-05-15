package projectsettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/projectservice"
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
		return apperrors.Wrap(err)
	}
	return nil
}

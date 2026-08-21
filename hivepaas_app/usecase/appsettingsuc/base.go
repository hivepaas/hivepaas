package appsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
)

const (
	defaultServiceRetryMax = 2
)

type persistingAppData struct {
	appservice.PersistingAppData
}

func (uc *UC) preparePersistingAppTags(
	app *entity.App,
	tags []string,
	startIndex int,
	persistingData *persistingAppData,
) {
	index := startIndex
	for _, tag := range tags {
		persistingData.UpsertingTags = append(persistingData.UpsertingTags,
			&entity.Tag{
				ObjectID: app.ID,
				Tag:      tag,
				Index:    index,
			})
		index++
	}
}

func (uc *UC) persistData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingAppData,
) error {
	err := uc.appService.PersistAppData(ctx, db, &persistingData.PersistingAppData)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
